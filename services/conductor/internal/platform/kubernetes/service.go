package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeitlos/lucity/pkg/to"
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

	replicaSetsByService := make(map[string][]apps.ReplicaSet)

	for _, replicaSet := range replicaSets.Items {
		name := replicaSet.Labels[serviceLabel]
		replicaSetsByService[name] = append(replicaSetsByService[name], replicaSet)
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
			routesByService[name],
			hpasByService[name],
			environmentID,
		)

		services = append(services, service)
	}

	return services, nil
}

func (c *Client) ServicesByRepo(ctx context.Context, repoURL, branch string) ([]platform.ServiceID, error) {
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

	var result []platform.ServiceID

	for _, deployment := range deployments.Items {
		if deployment.Annotations[annotationSourceRepo] != repoURL {
			continue
		}

		if branch != "" && deployment.Annotations[annotationSourceBranch] != branch {
			continue
		}

		result = append(result, serviceID(deployment, environmentID(deployment.Labels)))
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

	return new(toService(*deployment, replicaSets, routes.Items, hpa, id.EnvironmentID())), nil
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

func toService(deployment apps.Deployment, replicaSets []apps.ReplicaSet, routes []unstructured.Unstructured, hpa autoscaling.HorizontalPodAutoscaler, environmentID platform.EnvironmentID) platform.Service {
	annotations := deployment.Annotations
	containers := deployment.Spec.Template.Spec.Containers

	service := platform.Service{
		ID:   serviceID(deployment, environmentID),
		Name: deployment.Labels[serviceLabel],

		Status: serviceStatus(deployment, replicaSets),
		Replicas: platform.ReplicaCount{
			Desired: int(to.Val(deployment.Spec.Replicas)),
			Ready:   int(deployment.Status.ReadyReplicas),
		},
		Autoscaling: autoscalingSettings(hpa),

		Port:      containersPort(containers),
		Endpoints: endpoints(deployment, routes),

		SourceURL:   annotations[annotationSourceRepo],
		Branch:      annotations[annotationSourceBranch],
		ContextPath: annotations[annotationSourceContext],
		Resources:   containerResources(containers),
		Command:     containerCommand(containers),
		Variables:   containerVariables(containers),

		LastDeployedAt: latestReplicaSetTime(replicaSets),
		CreatedAt:      deployment.CreationTimestamp.Time,
	}

	currentRevision := deployment.Annotations[annotationRevision]

	for _, replicaSet := range replicaSets {
		if replicaSet.Annotations[annotationRevision] != currentRevision {
			continue
		}

		service.ActiveDeployment = new(toDeployment(replicaSet, deployment, service.ID))
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

func serviceStatus(deployment apps.Deployment, replicaSets []apps.ReplicaSet) platform.ServiceStatus {
	if deployment.Annotations[annotationAwaitingBuild] == "true" {
		return platform.ServiceBuilding
	}

	desired := to.Val(deployment.Spec.Replicas)

	if desired == 0 {
		return platform.ServiceStopped
	}

	currentRevision := deployment.Annotations[annotationRevision]

	// Rollout in flight if any non-current replica set still has live pods
	for _, replicaSet := range replicaSets {
		if replicaSet.Annotations[annotationRevision] == currentRevision {
			continue
		}

		if replicaSet.Status.Replicas > 0 {
			return platform.ServiceDeploying
		}
	}

	// Only the current replica set has pods, indicating a steady state.
	if rolloutFailed(deployment) {
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

// containerVariables returns the literal env vars set on the running
// container. Excludes HOST/PORT (chart defaults) and entries sourced from
// secrets or ConfigMaps (those represent refs, not user-set literals).
func containerVariables(containers []core.Container) map[string]string {
	if len(containers) == 0 {
		return nil
	}

	out := map[string]string{}

	for _, e := range containers[0].Env {
		if e.Name == "HOST" || e.Name == "PORT" {
			continue
		}

		if e.ValueFrom != nil {
			continue
		}

		out[e.Name] = e.Value
	}

	return out
}

// containerCommand returns the user's command override as a single string.
// Empty means no override (image default is used at runtime).
func containerCommand(containers []core.Container) string {
	if len(containers) == 0 {
		return ""
	}

	parts := append([]string{}, containers[0].Command...)
	parts = append(parts, containers[0].Args...)

	return strings.Join(parts, " ")
}

var httpRouteGVR = schema.GroupVersionResource{
	Group:    "gateway.networking.k8s.io",
	Version:  "v1",
	Resource: "httproutes",
}
