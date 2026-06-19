package kubernetes

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zeitlos/lucity/pkg/to"
	"github.com/zeitlos/lucity/services/conductor/internal/image"
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
	containers := replicaSet.Spec.Template.Spec.Containers
	annotations := replicaSet.Spec.Template.Annotations

	result := platform.Deployment{
		ID:     deploymentID(replicaSet, serviceID),
		Status: deploymentStatus(replicaSet, deployment),

		Commit:        annotations[annotationSourceCommit],
		CommitMessage: annotations[annotationSourceMessage],
		Ref:           annotations[annotationSourceRef],
		SourceURL:     annotations[annotationSourceRepo],
		ContextPath:   annotations[annotationSourceContext],

		Resources: containerResources(containers),
		Command:   containerCommand(containers),

		BuildID: annotations[annotationBuildID],

		Replicas: platform.ReplicaCount{
			Desired: int(to.Val(replicaSet.Spec.Replicas)),
			Ready:   int(replicaSet.Status.ReadyReplicas),
		},

		CreatedAt: replicaSet.CreationTimestamp.Time,
	}

	if len(containers) > 0 {
		if ref, err := image.Parse(containers[0].Image); err == nil {
			result.Image = ref
		}
	}

	if id, ok := deployment.Labels[gitHubInstallationLabel]; ok {
		if parsed, err := strconv.Atoi(id); err == nil {
			result.GitHubInstallationID = parsed
		}
	}

	return result
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
	currentRevision := deployment.Annotations[annotationRevision]

	if replicaSet.Annotations[annotationRevision] != currentRevision {
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
