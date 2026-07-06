package kubernetes

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/zeitlos/lucity/pkg/to"
	"github.com/zeitlos/lucity/services/conductor/internal/image"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
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

	pods, err := c.podsFor(ctx, serviceID)

	if err != nil {
		return nil, err
	}

	result := make([]platform.Deployment, 0, len(replicaSets))

	for _, replicaSet := range replicaSets {
		result = append(result, toDeployment(replicaSet, *deployment, pods, serviceID))
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

	pods, err := c.podsFor(ctx, id.ServiceID())

	if err != nil {
		return nil, err
	}

	return new(toDeployment(replicaSets.Items[0], *deployment, pods, id.ServiceID())), nil
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

func (c *Client) podsFor(ctx context.Context, serviceID platform.ServiceID) ([]core.Pod, error) {
	set := labels.Set{
		serviceLabel: serviceID.Name,
	}

	pods, err := c.kubernetes.CoreV1().Pods(serviceID.Namespace()).List(ctx, meta.ListOptions{
		LabelSelector: labels.SelectorFromSet(set).String(),
	})

	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func toDeployment(replicaSet apps.ReplicaSet, deployment apps.Deployment, pods []core.Pod, serviceID platform.ServiceID) platform.Deployment {
	containers := replicaSet.Spec.Template.Spec.Containers
	annotations := replicaSet.Spec.Template.Annotations
	rollout := rolloutFor(replicaSet, deployment, pods)

	result := platform.Deployment{
		ID:      deploymentID(replicaSet, serviceID),
		Status:  deploymentStatus(replicaSet, deployment, rollout),
		Rollout: rollout,

		Commit:        annotations[annotationSourceCommit],
		CommitMessage: annotations[annotationSourceMessage],
		Ref:           annotations[annotationSourceRef],
		SourceURL:     annotations[annotationSourceRepo],
		ContextPath:   annotations[annotationSourceContext],

		Resources: containerResources(containers),
		Command:   containerCommand(containers),

		BuildID:        annotations[annotationBuildID],
		ReleaseID:      annotations[annotationRelease],
		ReleaseTrigger: annotations[annotationReleaseTrigger],
		ReleaseActor:   annotations[annotationReleaseActor],

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

func deploymentStatus(replicaSet apps.ReplicaSet, deployment apps.Deployment, rollout *platform.Rollout) platform.DeploymentStatus {
	if replicaSet.Annotations[annotationRevision] != deployment.Annotations[annotationRevision] {
		return platform.DeploymentSuperseded
	}

	if rollout == nil {
		return platform.DeploymentActive
	}

	switch rollout.Status {
	case platform.RolloutReady:
		return platform.DeploymentActive
	case platform.RolloutFailed:
		return platform.DeploymentFailed
	case platform.RolloutDegraded:
		if rolloutCompleted(deployment) {
			return platform.DeploymentActive
		}

		return platform.DeploymentDeploying
	}

	return platform.DeploymentDeploying
}

func rolloutFor(replicaSet apps.ReplicaSet, deployment apps.Deployment, pods []core.Pod) *platform.Rollout {
	containers := replicaSet.Spec.Template.Spec.Containers

	if len(containers) == 0 {
		return nil
	}

	ref, err := image.Parse(containers[0].Image)

	if err != nil || !ref.Built() {
		return nil
	}

	rollout := platform.Rollout{StartedAt: replicaSet.CreationTimestamp.Time}

	if replicaSet.Annotations[annotationRevision] != deployment.Annotations[annotationRevision] {
		rollout.Status = platform.RolloutSuperseded

		return &rollout
	}

	revisionPods := podsOfRevision(pods, replicaSet.Labels[podTemplateHashLabel])

	for _, pod := range revisionPods {
		for _, container := range pod.Status.ContainerStatuses {
			rollout.Restarts += int(container.RestartCount)
		}
	}

	desired := to.Val(replicaSet.Spec.Replicas)
	ready := replicaSet.Status.ReadyReplicas
	reason, message := diagnoseReplicaSet(replicaSet)

	if reason == platform.RolloutReasonNone {
		reason, message = diagnosePods(revisionPods)
	}

	switch {
	case desired == 0 || ready == desired:
		rollout.Status = platform.RolloutReady

	case reason != platform.RolloutReasonNone:
		rollout.Reason = reason
		rollout.Message = message
		rollout.Status = platform.RolloutDegraded

		if ready == 0 {
			rollout.Status = platform.RolloutFailed
		}

	case rolloutFailed(deployment):
		rollout.Reason = platform.RolloutReasonDeadlineExceeded
		rollout.Message = "rollout made no progress"
		rollout.Status = platform.RolloutDegraded

		if ready == 0 {
			rollout.Status = platform.RolloutFailed

			if anyContainerRunning(revisionPods) {
				rollout.Reason = platform.RolloutReasonNotReady
				rollout.Message = "container is running but never became ready"
			}
		}

	default:
		rollout.Status = platform.RolloutProgressing
	}

	return &rollout
}

func podsOfRevision(pods []core.Pod, podTemplateHash string) []core.Pod {
	result := make([]core.Pod, 0, len(pods))

	for _, pod := range pods {
		if pod.Labels[podTemplateHashLabel] == podTemplateHash {
			result = append(result, pod)
		}
	}

	return result
}

func diagnoseReplicaSet(replicaSet apps.ReplicaSet) (platform.RolloutReason, string) {
	for _, condition := range replicaSet.Status.Conditions {
		if condition.Type != apps.ReplicaSetReplicaFailure || condition.Status != core.ConditionTrue {
			continue
		}

		if strings.Contains(condition.Message, "exceeded quota") {
			return platform.RolloutReasonQuotaExceeded, "workspace resource limit reached"
		}

		return platform.RolloutReasonConfigError, "pods could not be created"
	}

	return platform.RolloutReasonNone, ""
}

func diagnosePods(pods []core.Pod) (platform.RolloutReason, string) {
	for _, pod := range pods {
		if podReady(pod) {
			continue
		}

		for _, container := range pod.Status.ContainerStatuses {
			if oomKilled(container) {
				return platform.RolloutReasonOOMKilled, "container was killed: out of memory"
			}
		}
	}

	for _, pod := range pods {
		if podReady(pod) {
			continue
		}

		for _, container := range pod.Status.ContainerStatuses {
			waiting := container.State.Waiting

			if waiting == nil {
				continue
			}

			switch waiting.Reason {
			case "CrashLoopBackOff":
				if terminated := container.LastTerminationState.Terminated; terminated != nil {
					return platform.RolloutReasonCrashLoop, fmt.Sprintf("container exited with code %d", terminated.ExitCode)
				}

				return platform.RolloutReasonCrashLoop, "container is restarting repeatedly"
			case "ImagePullBackOff", "ErrImagePull", "InvalidImageName":
				return platform.RolloutReasonImagePullFailed, "image could not be pulled"
			case "CreateContainerConfigError", "CreateContainerError":
				return platform.RolloutReasonConfigError, "container could not be created from its configuration"
			}
		}

		for _, condition := range pod.Status.Conditions {
			if condition.Type == core.PodScheduled && condition.Status == core.ConditionFalse && condition.Reason == core.PodReasonUnschedulable {
				return platform.RolloutReasonUnschedulable, "no capacity available to schedule"
			}
		}
	}

	return platform.RolloutReasonNone, ""
}

func podReady(pod core.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == core.PodReady {
			return condition.Status == core.ConditionTrue
		}
	}

	return false
}

func oomKilled(container core.ContainerStatus) bool {
	if terminated := container.State.Terminated; terminated != nil && terminated.Reason == "OOMKilled" {
		return true
	}

	if terminated := container.LastTerminationState.Terminated; terminated != nil && terminated.Reason == "OOMKilled" {
		return true
	}

	return false
}

func anyContainerRunning(pods []core.Pod) bool {
	for _, pod := range pods {
		for _, container := range pod.Status.ContainerStatuses {
			if container.State.Running != nil {
				return true
			}
		}
	}

	return false
}

func rolloutCompleted(deployment apps.Deployment) bool {
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == apps.DeploymentProgressing {
			return condition.Status == core.ConditionTrue && condition.Reason == "NewReplicaSetAvailable"
		}
	}

	return false
}
