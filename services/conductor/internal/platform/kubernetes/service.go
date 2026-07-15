package kubernetes

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zeitlos/lucity/pkg/to"
	"github.com/zeitlos/lucity/services/conductor/internal/image"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	apps "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v2"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
)

const portName = "http"

func (c *Client) Services(ctx context.Context, environmentID platform.EnvironmentID) ([]platform.Service, error) {
	req, err := labels.NewRequirement(serviceLabel, selection.Exists, nil)

	if err != nil {
		return nil, err
	}

	selector := labels.NewSelector().Add(*req)
	namespace := environmentID.Namespace()

	deployments, err := c.kubernetes.AppsV1().Deployments(namespace).List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	replicaSets, err := c.kubernetes.AppsV1().ReplicaSets(namespace).List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	secrets, err := c.kubernetes.CoreV1().Secrets(namespace).List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	routes, err := c.dynamic.Resource(httpRouteGVR).Namespace(namespace).List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	hpas, err := c.kubernetes.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	pods, err := c.kubernetes.CoreV1().Pods(namespace).List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	replicaSetsByService := make(map[string][]apps.ReplicaSet)

	for _, replicaSet := range replicaSets.Items {
		name := replicaSet.Labels[serviceLabel]
		replicaSetsByService[name] = append(replicaSetsByService[name], replicaSet)
	}

	podsByService := make(map[string][]core.Pod)

	for _, pod := range pods.Items {
		name := pod.Labels[serviceLabel]
		podsByService[name] = append(podsByService[name], pod)
	}

	secretsByService := make(map[string]core.Secret)

	for _, secret := range secrets.Items {
		name := secret.Labels[serviceLabel]
		secretsByService[name] = secret
	}

	routesByService := make(map[string][]unstructured.Unstructured)

	for _, route := range routes.Items {
		name := route.GetLabels()[serviceLabel]
		routesByService[name] = append(routesByService[name], route)
	}

	hpasByService := make(map[string]autoscaling.HorizontalPodAutoscaler)

	for _, hpa := range hpas.Items {
		hpasByService[hpa.Labels[serviceLabel]] = hpa
	}

	services := make([]platform.Service, 0, len(deployments.Items))

	for _, deployment := range deployments.Items {
		name := deployment.Labels[serviceLabel]
		service := toService(
			deployment,
			replicaSetsByService[name],
			podsByService[name],
			routesByService[name],
			hpasByService[name],
			secretsByService[name],
			environmentID,
		)

		services = append(services, service)
	}

	return services, nil
}

func (c *Client) ServicesByRepo(ctx context.Context, repoURL string) ([]platform.RepoService, error) {
	req, err := labels.NewRequirement(serviceLabel, selection.Exists, nil)

	if err != nil {
		return nil, err
	}

	selector := labels.NewSelector().Add(*req)

	deployments, err := c.kubernetes.AppsV1().Deployments("").List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	var result []platform.RepoService

	wantRepo := strings.TrimSuffix(repoURL, ".git")

	for _, deployment := range deployments.Items {
		if !strings.EqualFold(strings.TrimSuffix(deployment.Annotations[annotationSourceRepo], ".git"), wantRepo) {
			continue
		}

		installationID, _ := strconv.ParseInt(deployment.Labels[gitHubInstallationLabel], 10, 64)

		result = append(result, platform.RepoService{
			ID:             serviceID(deployment, environmentID(deployment.Labels)),
			Branch:         deployment.Annotations[annotationSourceBranch],
			AutoDeploy:     deployment.Annotations[annotationAutoDeploy] == "true",
			CIDeploy:       deployment.Annotations[annotationCIDeploy] == "true",
			InstallationID: installationID,
		})
	}

	return result, nil
}

func (c *Client) Service(ctx context.Context, id platform.ServiceID) (*platform.Service, error) {
	deployment, err := c.deploymentFor(ctx, id)

	if err != nil {
		return nil, err
	}

	replicaSets, err := c.replicaSetsFor(ctx, id)

	if err != nil {
		return nil, err
	}

	pods, err := c.podsFor(ctx, id)

	if err != nil {
		return nil, err
	}

	secret, err := c.secretFor(ctx, id)

	if err != nil {
		return nil, err
	}

	selector := labels.SelectorFromSet(labels.Set{serviceLabel: id.Name}).String()

	routes, err := c.dynamic.Resource(httpRouteGVR).Namespace(id.Namespace()).List(ctx, meta.ListOptions{
		LabelSelector: selector,
	})

	if err != nil {
		return nil, err
	}

	hpas, err := c.kubernetes.AutoscalingV2().HorizontalPodAutoscalers(id.Namespace()).List(ctx, meta.ListOptions{
		LabelSelector: selector,
	})

	if err != nil {
		return nil, err
	}

	var hpa autoscaling.HorizontalPodAutoscaler

	if len(hpas.Items) > 0 {
		hpa = hpas.Items[0]
	}

	return new(toService(*deployment, replicaSets, pods, routes.Items, hpa, *secret, id.EnvironmentID())), nil
}

func (c *Client) deploymentFor(ctx context.Context, serviceID platform.ServiceID) (*apps.Deployment, error) {
	set := labels.Set{
		serviceLabel: serviceID.Name,
	}

	deployments, err := c.kubernetes.AppsV1().Deployments(serviceID.Namespace()).List(ctx, meta.ListOptions{
		LabelSelector: labels.SelectorFromSet(set).String(),
	})

	if err != nil {
		return nil, err
	}

	if len(deployments.Items) == 0 {
		return nil, fmt.Errorf("service %q not found", serviceID)
	}

	return &deployments.Items[0], nil
}

func (c *Client) secretFor(ctx context.Context, serviceID platform.ServiceID) (*core.Secret, error) {
	set := labels.Set{
		serviceLabel: serviceID.Name,
	}

	secrets, err := c.kubernetes.CoreV1().Secrets(serviceID.Namespace()).List(ctx, meta.ListOptions{
		LabelSelector: labels.SelectorFromSet(set).String(),
	})

	if err != nil {
		return nil, err
	}

	if len(secrets.Items) == 0 {
		return nil, fmt.Errorf("variables for service %q not found", serviceID)
	}

	return &secrets.Items[0], nil
}

func toService(deployment apps.Deployment, replicaSets []apps.ReplicaSet, pods []core.Pod, routes []unstructured.Unstructured, hpa autoscaling.HorizontalPodAutoscaler, secret core.Secret, environmentID platform.EnvironmentID) platform.Service {
	annotations := deployment.Annotations
	containers := deployment.Spec.Template.Spec.Containers

	id := serviceID(deployment, environmentID)
	currentRevision := deployment.Annotations[annotationRevision]

	deployments := make([]platform.Deployment, 0, len(replicaSets))

	var activeDeployment *platform.Deployment

	for i, replicaSet := range replicaSets {
		deployments = append(deployments, toDeployment(replicaSet, deployment, pods, id))

		if replicaSet.Annotations[annotationRevision] == currentRevision {
			activeDeployment = &deployments[i]
		}
	}

	var rollout *platform.Rollout

	if activeDeployment != nil {
		rollout = activeDeployment.Rollout
	}

	service := platform.Service{
		ID:   id,
		Name: deployment.Labels[serviceLabel],

		Status: serviceStatus(deployment, replicaSets, rollout),
		Replicas: platform.ReplicaCount{
			Desired: int(to.Val(deployment.Spec.Replicas)),
			Ready:   int(deployment.Status.ReadyReplicas),
		},
		Autoscaling: autoscalingSettings(hpa),

		Port:      containersPort(containers),
		Endpoints: endpoints(deployment, routes),

		SourceURL:   annotations[annotationSourceRepo],
		Branch:      annotations[annotationSourceBranch],
		AutoDeploy:  annotations[annotationAutoDeploy] == "true",
		CIDeploy:    annotations[annotationCIDeploy] == "true",
		ContextPath: annotations[annotationSourceContext],
		Resources:   containerResources(containers),
		Command:     containerCommand(containers),
		HealthCheck: containerHealthCheck(containers),
		Variables:   make(map[string]string),

		ActiveDeployment: activeDeployment,
		Deployments:      deployments,

		LastDeployedAt: latestReplicaSetTime(replicaSets),
		CreatedAt:      deployment.CreationTimestamp.Time,
	}

	for key, val := range secret.Data {
		service.Variables[key] = string(val)
	}

	return service
}

func serviceID(deployment apps.Deployment, environmentID platform.EnvironmentID) platform.ServiceID {
	return platform.ServiceID{
		Workspace:   environmentID.Workspace,
		Project:     environmentID.Project,
		Environment: environmentID.Name,
		Name:        deployment.Labels[serviceLabel],
	}
}

func awaitingFirstBuild(deployment apps.Deployment) bool {
	containers := deployment.Spec.Template.Spec.Containers

	if len(containers) == 0 {
		return false
	}

	ref, err := image.Parse(containers[0].Image)

	return err == nil && !ref.Built()
}

func serviceStatus(deployment apps.Deployment, replicaSets []apps.ReplicaSet, rollout *platform.Rollout) platform.ServiceStatus {
	if awaitingFirstBuild(deployment) {
		return platform.ServiceBuilding
	}

	desired := to.Val(deployment.Spec.Replicas)

	if desired == 0 {
		return platform.ServiceStopped
	}

	rolloutHasFailed := rollout != nil && rollout.Status == platform.RolloutFailed
	currentRevision := deployment.Annotations[annotationRevision]

	// Rollout in flight if any non-current replica set still has live pods
	for _, replicaSet := range replicaSets {
		if replicaSet.Annotations[annotationRevision] == currentRevision {
			continue
		}

		if replicaSet.Status.Replicas > 0 {
			if rolloutHasFailed {
				return platform.ServiceDegraded
			}

			return platform.ServiceDeploying
		}
	}

	// Only the current replica set has pods, indicating a steady state.
	if rolloutHasFailed {
		return platform.ServiceFailed
	}

	if deployment.Status.ReadyReplicas == 0 {
		if rolloutProgressing(deployment) {
			return platform.ServiceDeploying
		}

		return platform.ServiceFailed
	}

	if deployment.Status.ReadyReplicas != desired {
		return platform.ServiceDegraded
	}

	return platform.ServiceHealthy
}

func rolloutFailed(deployment apps.Deployment) bool {
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == apps.DeploymentProgressing &&
			condition.Status == core.ConditionFalse &&
			condition.Reason == "ProgressDeadlineExceeded" {
			return true
		}
	}

	return false
}

func rolloutProgressing(deployment apps.Deployment) bool {
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == apps.DeploymentProgressing {
			return condition.Status == core.ConditionTrue
		}
	}

	return false
}

func latestReplicaSetTime(replicaSets []apps.ReplicaSet) time.Time {
	var latest time.Time

	for _, replicaSet := range replicaSets {
		if replicaSet.CreationTimestamp.After(latest) {
			latest = replicaSet.CreationTimestamp.Time
		}
	}

	return latest
}

func endpoints(deployment apps.Deployment, routes []unstructured.Unstructured) []platform.Endpoint {
	var port int
	containers := deployment.Spec.Template.Spec.Containers

	if len(containers) > 0 && len(containers[0].Ports) > 0 {
		port = int(containers[0].Ports[0].ContainerPort)
	}

	var endpoints []platform.Endpoint

	// TODO: Fetch the k8s service resource to derive the proper internal host name
	if port > 0 {
		endpoints = append(endpoints, platform.Endpoint{
			Host:     fmt.Sprintf("%s.%s.svc.cluster.local", deployment.Name, deployment.Namespace),
			Port:     port,
			Protocol: platform.ProtocolTCP,
		})
	}

	for _, route := range routes {
		hosts, _, _ := unstructured.NestedStringSlice(route.Object, "spec", "hostnames")
		parentRefs, _, _ := unstructured.NestedSlice(route.Object, "spec", "parentRefs")

		enabled := len(parentRefs) > 0

		for _, host := range hosts {
			endpoints = append(endpoints, platform.Endpoint{
				Enabled:  enabled,
				Host:     host,
				Port:     443,
				Protocol: platform.ProtocolHTTPS,
			})
		}
	}

	return endpoints
}

func autoscalingSettings(hpa autoscaling.HorizontalPodAutoscaler) *platform.AutoscalingSettings {
	if hpa.Name == "" {
		return nil
	}

	settings := &platform.AutoscalingSettings{
		MinReplicas: int(to.Val(hpa.Spec.MinReplicas)),
		MaxReplicas: int(hpa.Spec.MaxReplicas),
	}

	for _, metric := range hpa.Spec.Metrics {
		if metric.Type != autoscaling.ResourceMetricSourceType {
			continue
		}
		if metric.Resource == nil || metric.Resource.Name != core.ResourceCPU {
			continue
		}
		if metric.Resource.Target.AverageUtilization != nil {
			settings.TargetCPU = int(*metric.Resource.Target.AverageUtilization)
		}
	}

	return settings
}

func containersPort(containers []core.Container) int {
	if len(containers) == 0 {
		return 0
	}

	for _, port := range containers[0].Ports {
		if port.Name != portName {
			continue
		}

		return int(port.ContainerPort)
	}

	return 0
}

func containerResources(containers []core.Container) platform.Resources {
	if len(containers) == 0 {
		return platform.Resources{}
	}

	limits := containers[0].Resources.Limits

	return platform.Resources{
		CPU:    limits[core.ResourceCPU],
		Memory: limits[core.ResourceMemory],
	}
}

func containerCommand(containers []core.Container) string {
	if len(containers) == 0 {
		return ""
	}

	command := containers[0].Command

	if len(command) == 2 && command[0] == "/bin/sh" && command[1] == "-c" {
		return strings.Join(containers[0].Args, " ")
	}

	return ""
}

// containerHealthCheck reconstructs the health-check config from the rendered
// probes. Only an httpGet readiness probe maps to a configured health check;
// the default tcpSocket probe reads back as no health check (nil).
func containerHealthCheck(containers []core.Container) *platform.HealthCheck {
	if len(containers) == 0 {
		return nil
	}

	probe := containers[0].ReadinessProbe

	if probe == nil || probe.HTTPGet == nil {
		return nil
	}

	healthCheck := &platform.HealthCheck{
		Path:                probe.HTTPGet.Path,
		Port:                probe.HTTPGet.Port.IntValue(),
		InitialDelaySeconds: int(probe.InitialDelaySeconds),
		PeriodSeconds:       int(probe.PeriodSeconds),
		TimeoutSeconds:      int(probe.TimeoutSeconds),
		FailureThreshold:    int(probe.FailureThreshold),
	}

	if startup := containers[0].StartupProbe; startup != nil && startup.HTTPGet != nil {
		healthCheck.StartupFailureThreshold = int(startup.FailureThreshold)
	}

	return healthCheck
}

var httpRouteGVR = schema.GroupVersionResource{
	Group:    "gateway.networking.k8s.io",
	Version:  "v1",
	Resource: "httproutes",
}
