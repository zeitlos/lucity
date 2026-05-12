package kubernetes

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/pkg/to"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

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

	byService := make(map[string][]apps.ReplicaSet)

	for _, replicaSet := range replicaSets.Items {
		byService[replicaSet.Labels[serviceLabel]] = append(byService[replicaSet.Labels[serviceLabel]], replicaSet)
	}

	services := make([]platform.Service, 0, len(deployments.Items))

	for _, deployment := range deployments.Items {
		services = append(services, toService(deployment, byService[deployment.Labels[serviceLabel]], environmentID))
	}

	return services, nil
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

	return new(toService(*deployment, replicaSets, id.EnvironmentID())), nil
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

func toService(deployment apps.Deployment, replicaSets []apps.ReplicaSet, environmentID platform.EnvironmentID) platform.Service {
	return platform.Service{
		ID:   serviceID(deployment, environmentID),
		Name: deployment.Labels[serviceLabel],

		Status: serviceStatus(deployment, replicaSets),

		CreatedAt: deployment.CreationTimestamp.Time,
	}
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
	desired := to.Val(deployment.Spec.Replicas)

	if desired == 0 {
		return platform.ServiceStopped
	}

	currentHash := deployment.Spec.Template.Labels[podTemplateHashLabel]

	// Rollout in flight if any non-current replica set still has live pods
	for _, replicaSet := range replicaSets {
		if replicaSet.Labels[podTemplateHashLabel] == currentHash {
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
