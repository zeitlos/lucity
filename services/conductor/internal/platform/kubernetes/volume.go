package kubernetes

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const volumeMountPrefix = "volume-"

func (c *Client) Volumes(ctx context.Context, environmentID platform.EnvironmentID) ([]platform.Volume, error) {
	req, err := labels.NewRequirement(volumeLabel, selection.Exists, nil)

	if err != nil {
		return nil, err
	}

	selector := labels.NewSelector().Add(*req)

	list, err := c.kubernetes.CoreV1().PersistentVolumeClaims(environmentID.Namespace()).List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	mounts, err := c.volumeMounts(ctx, environmentID)

	if err != nil {
		return nil, err
	}

	volumes := make([]platform.Volume, 0, len(list.Items))

	for _, pvc := range list.Items {
		volume := toVolume(pvc, environmentID)

		if mount, ok := mounts[volume.Name]; ok {
			volume.Mount = &mount
		}

		volumes = append(volumes, volume)
	}

	return volumes, nil
}

func (c *Client) Volume(ctx context.Context, id platform.VolumeID) (*platform.Volume, error) {
	set := labels.Set{
		volumeLabel: id.Name,
	}

	list, err := c.kubernetes.CoreV1().PersistentVolumeClaims(id.Namespace()).List(ctx, meta.ListOptions{
		LabelSelector: labels.SelectorFromSet(set).String(),
	})

	if err != nil {
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("volume %q not found", id)
	}

	volume := toVolume(list.Items[0], id.EnvironmentID())

	mounts, err := c.volumeMounts(ctx, id.EnvironmentID())

	if err != nil {
		return nil, err
	}

	if mount, ok := mounts[volume.Name]; ok {
		volume.Mount = &mount
	}

	return &volume, nil
}

func (c *Client) volumeMounts(ctx context.Context, environmentID platform.EnvironmentID) (map[string]platform.VolumeMount, error) {
	req, err := labels.NewRequirement(serviceLabel, selection.Exists, nil)

	if err != nil {
		return nil, err
	}

	selector := labels.NewSelector().Add(*req)

	deployments, err := c.kubernetes.AppsV1().Deployments(environmentID.Namespace()).List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	mounts := make(map[string]platform.VolumeMount)

	for _, deployment := range deployments.Items {
		service := serviceID(deployment, environmentID)

		for volumeName, path := range deploymentVolumeMounts(deployment) {
			mounts[volumeName] = platform.VolumeMount{
				Service: service,
				Path:    path,
			}
		}
	}

	return mounts, nil
}

func deploymentVolumeMounts(deployment apps.Deployment) map[string]string {
	spec := deployment.Spec.Template.Spec

	paths := make(map[string]string)

	for _, container := range spec.Containers {
		for _, mount := range container.VolumeMounts {
			paths[mount.Name] = mount.MountPath
		}
	}

	mounts := make(map[string]string)

	for _, volume := range spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}

		if !strings.HasPrefix(volume.Name, volumeMountPrefix) {
			continue
		}

		mounts[strings.TrimPrefix(volume.Name, volumeMountPrefix)] = paths[volume.Name]
	}

	return mounts
}

func toVolume(pvc core.PersistentVolumeClaim, environmentID platform.EnvironmentID) platform.Volume {
	return platform.Volume{
		ID:        volumeID(pvc, environmentID),
		Name:      pvc.Labels[volumeLabel],
		Size:      pvc.Spec.Resources.Requests[core.ResourceStorage],
		Status:    volumeStatus(pvc),
		CreatedAt: pvc.CreationTimestamp.Time,
	}
}

func volumeID(pvc core.PersistentVolumeClaim, environmentID platform.EnvironmentID) platform.VolumeID {
	return platform.VolumeID{
		Workspace:   environmentID.Workspace,
		Project:     environmentID.Project,
		Environment: environmentID.Name,
		Name:        pvc.Labels[volumeLabel],
	}
}

func volumeStatus(pvc core.PersistentVolumeClaim) platform.VolumeStatus {
	switch pvc.Status.Phase {
	case core.ClaimBound:
		return platform.VolumeReady
	case core.ClaimLost:
		return platform.VolumeFailed
	default:
		return platform.VolumePending
	}
}
