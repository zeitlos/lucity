package kubernetes

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/pkg/to"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
	apps "k8s.io/api/apps/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Client) Deployments(ctx context.Context, serviceID platform.ServiceID) ([]platform.Deployment, error) {
	deployment, err := c.deploymentFor(ctx, serviceID)

	if err != nil {
		return nil, err
	}

	replicaSets, err := c.replicaSetsFor(ctx, serviceID)

	if err != nil {
		return nil, err
	}

	result := make([]platform.Deployment, 0, len(replicaSets))

	for _, replicaSet := range replicaSets {
		result = append(result, toDeployment(replicaSet, *deployment, serviceID))
	}

	return result, nil
}

func (c *Client) Deployment(ctx context.Context, id platform.DeploymentID) (*platform.Deployment, error) {
	set := labels.Set{
		serviceLabel:         id.Service,
		podTemplateHashLabel: id.Hash,
	}

	selector := labels.SelectorFromSet(set)
	namespace := id.Namespace()

	replicaSets, err := c.kubernetes.AppsV1().ReplicaSets(namespace).List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	if len(replicaSets.Items) == 0 {
		return nil, fmt.Errorf("deployment %q not found", id)
	}

	deployment, err := c.deploymentFor(ctx, id.ServiceID())

	if err != nil {
		return nil, err
	}

	return new(toDeployment(replicaSets.Items[0], *deployment, id.ServiceID())), nil
}

func (c *Client) replicaSetsFor(ctx context.Context, serviceID platform.ServiceID) ([]apps.ReplicaSet, error) {
	set := labels.Set{
		serviceLabel: serviceID.Name,
	}

	replicaSets, err := c.kubernetes.AppsV1().ReplicaSets(serviceID.Namespace()).List(ctx, meta.ListOptions{
		LabelSelector: labels.SelectorFromSet(set).String(),
	})

	if err != nil {
		return nil, err
	}

	return replicaSets.Items, nil
}

func toDeployment(replicaSet apps.ReplicaSet, deployment apps.Deployment, serviceID platform.ServiceID) platform.Deployment {
	image := ""

	if len(replicaSet.Spec.Template.Spec.Containers) > 0 {
		image = replicaSet.Spec.Template.Spec.Containers[0].Image
	}

	return platform.Deployment{
		ID:    deploymentID(replicaSet, serviceID),
		Image: image,

		Status: deploymentStatus(replicaSet, deployment),
	}
}

func deploymentID(replicaSet apps.ReplicaSet, serviceID platform.ServiceID) platform.DeploymentID {
	return platform.DeploymentID{
		Workspace:   serviceID.Workspace,
		Project:     serviceID.Project,
		Environment: serviceID.Environment,
		Service:     serviceID.Name,
		Hash:        replicaSet.Labels[podTemplateHashLabel],
	}
}

func deploymentStatus(replicaSet apps.ReplicaSet, deployment apps.Deployment) platform.DeploymentStatus {
	currentHash := deployment.Spec.Template.Labels[podTemplateHashLabel]

	if replicaSet.Labels[podTemplateHashLabel] != currentHash {
		return platform.DeploymentSuperseded
	}

	if rolloutFailed(deployment) {
		return platform.DeploymentFailed
	}

	desired := to.Val(replicaSet.Spec.Replicas)

	if desired == 0 || replicaSet.Status.ReadyReplicas == desired {
		return platform.DeploymentActive
	}

	return platform.DeploymentDeploying
}
