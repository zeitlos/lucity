package kubernetes

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (c *Client) KeyValueStores(ctx context.Context, environmentID platform.EnvironmentID) ([]platform.KeyValueStore, error) {
	req, err := labels.NewRequirement(keyValueStoreLabel, selection.Exists, nil)

	if err != nil {
		return nil, err
	}

	selector := labels.NewSelector().Add(*req)

	list, err := c.kubernetes.AppsV1().StatefulSets(environmentID.Namespace()).List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	stores := make([]platform.KeyValueStore, 0, len(list.Items))

	for _, item := range list.Items {
		stores = append(stores, toKeyValueStore(item, environmentID))
	}

	return stores, nil
}

func (c *Client) KeyValueStore(ctx context.Context, id platform.KeyValueStoreID) (*platform.KeyValueStore, error) {
	set := labels.Set{
		keyValueStoreLabel: id.Name,
	}

	list, err := c.kubernetes.AppsV1().StatefulSets(id.Namespace()).List(ctx, meta.ListOptions{
		LabelSelector: labels.SelectorFromSet(set).String(),
	})

	if err != nil {
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("key-value store %q not found", id)
	}

	return new(toKeyValueStore(list.Items[0], id.EnvironmentID())), nil
}

func (c *Client) KeyValueStoreCredentials(ctx context.Context, id platform.KeyValueStoreID) (*platform.KeyValueStoreCredentials, error) {
	secretName := "lucity-app-valkey-" + id.Name
	namespace := id.Namespace()

	secret, err := c.kubernetes.CoreV1().Secrets(namespace).Get(ctx, secretName, meta.GetOptions{})

	if apierrors.IsNotFound(err) {
		return nil, platform.ErrDatabaseProvisioning
	}

	if err != nil {
		return nil, fmt.Errorf("get key-value store secret %q in %q: %w", secretName, namespace, err)
	}

	return &platform.KeyValueStoreCredentials{
		Host:     string(secret.Data["host"]),
		Port:     string(secret.Data["port"]),
		Password: string(secret.Data["password"]),
		URI:      string(secret.Data["uri"]),
	}, nil
}

func toKeyValueStore(sts apps.StatefulSet, environmentID platform.EnvironmentID) platform.KeyValueStore {
	store := platform.KeyValueStore{
		ID:        keyValueStoreID(sts, environmentID),
		Name:      sts.Labels[keyValueStoreLabel],
		Status:    keyValueStoreStatus(sts),
		CreatedAt: sts.GetCreationTimestamp().Time,
	}

	if len(sts.Spec.Template.Spec.Containers) > 0 {
		// Image: "valkey/valkey:8-alpine"
		image := sts.Spec.Template.Spec.Containers[0].Image

		if i := strings.LastIndex(image, ":"); i != -1 {
			version := image[i+1:]

			if j := strings.Index(version, "-"); j != -1 {
				version = version[:j]
			}

			store.Version = version
		}
	}

	if len(sts.Spec.VolumeClaimTemplates) > 0 {
		if size, ok := sts.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[core.ResourceStorage]; ok {
			store.Size = size
		}
	}

	return store
}

func keyValueStoreID(sts apps.StatefulSet, environmentID platform.EnvironmentID) platform.KeyValueStoreID {
	return platform.KeyValueStoreID{
		Workspace:   environmentID.Workspace,
		Project:     environmentID.Project,
		Environment: environmentID.Name,
		Name:        sts.Labels[keyValueStoreLabel],
	}
}

func keyValueStoreStatus(sts apps.StatefulSet) platform.DatabaseStatus {
	desired := int32(1)

	if sts.Spec.Replicas != nil {
		desired = *sts.Spec.Replicas
	}

	ready := sts.Status.ReadyReplicas

	if desired == 0 {
		return platform.DatabaseStopped
	}

	if ready == 0 {
		return platform.DatabasePending
	}

	if ready < desired {
		return platform.DatabaseDegraded
	}

	return platform.DatabaseHealthy
}
