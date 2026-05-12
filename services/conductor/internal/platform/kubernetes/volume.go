package kubernetes

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

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

	volumes := make([]platform.Volume, 0, len(list.Items))

	for _, pvc := range list.Items {
		volumes = append(volumes, toVolume(pvc, environmentID))
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

	return new(toVolume(list.Items[0], id.EnvironmentID())), nil
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
