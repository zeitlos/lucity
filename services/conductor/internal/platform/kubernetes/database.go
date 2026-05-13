package kubernetes

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
)

var cnpgClusterGVR = cnpgv1.SchemeGroupVersion.WithResource("clusters")

func (c *Client) Databases(ctx context.Context, environmentID platform.EnvironmentID) ([]platform.Database, error) {
	req, err := labels.NewRequirement(databaseLabel, selection.Exists, nil)

	if err != nil {
		return nil, err
	}

	selector := labels.NewSelector().Add(*req)

	list, err := c.dynamic.Resource(cnpgClusterGVR).Namespace(environmentID.Namespace()).List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	databases := make([]platform.Database, 0, len(list.Items))

	for _, item := range list.Items {
		cluster, err := toCluster(item)

		if err != nil {
			return nil, err
		}

		databases = append(databases, toDatabase(*cluster, environmentID))
	}

	return databases, nil
}

func (c *Client) Database(ctx context.Context, id platform.DatabaseID) (*platform.Database, error) {
	set := labels.Set{
		databaseLabel: id.Name,
	}

	list, err := c.dynamic.Resource(cnpgClusterGVR).Namespace(id.Namespace()).List(ctx, meta.ListOptions{
		LabelSelector: labels.SelectorFromSet(set).String(),
	})

	if err != nil {
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("database %q not found", id)
	}

	cluster, err := toCluster(list.Items[0])

	if err != nil {
		return nil, err
	}

	return new(toDatabase(*cluster, id.EnvironmentID())), nil
}

func toCluster(item unstructured.Unstructured) (*cnpgv1.Cluster, error) {
	var cluster cnpgv1.Cluster
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &cluster)
	return &cluster, err
}

func toDatabase(cluster cnpgv1.Cluster, environmentID platform.EnvironmentID) platform.Database {
	database := platform.Database{
		ID:        databaseID(cluster, environmentID),
		Name:      cluster.Labels[databaseLabel],
		Instances: cluster.Spec.Instances,
		Status:    databaseStatus(cluster),
		CreatedAt: cluster.GetCreationTimestamp().Time,
	}

	// CNPG image: "ghcr.io/cloudnative-pg/postgresql:16.0"
	if i := strings.LastIndex(cluster.Spec.ImageName, ":"); i != -1 {
		database.Version = cluster.Spec.ImageName[i+1:]
	}

	size := cluster.Spec.StorageConfiguration.Size

	if size != "" {
		sizeQuantity, err := resource.ParseQuantity(size)

		if err != nil {
			slog.Warn("failed to parse size", "error", err, "database", cluster.Name, "namespace", cluster.Namespace)
		} else {
			database.Size = sizeQuantity
		}
	}

	return database
}

func databaseID(cluster cnpgv1.Cluster, environmentID platform.EnvironmentID) platform.DatabaseID {
	return platform.DatabaseID{
		Workspace:   environmentID.Workspace,
		Project:     environmentID.Project,
		Environment: environmentID.Name,
		Name:        cluster.Labels[databaseLabel],
	}
}

func databaseStatus(cluster cnpgv1.Cluster) platform.DatabaseStatus {
	desired := cluster.Spec.Instances
	ready := cluster.Status.ReadyInstances

	if desired == 0 {
		return platform.DatabaseStopped
	}

	if ready == 0 {
		isInitialBootstrap := cluster.Status.CurrentPrimary == ""

		if isInitialBootstrap {
			return platform.DatabasePending
		}

		return platform.DatabaseFailed
	}

	if ready < desired {
		return platform.DatabaseDegraded
	}

	return platform.DatabaseHealthy
}
