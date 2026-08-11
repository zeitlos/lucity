package kubernetes

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	cluster, err := c.cluster(ctx, id)

	if err != nil {
		return nil, err
	}

	return new(toDatabase(*cluster, id.EnvironmentID())), nil
}

func (c *Client) cluster(ctx context.Context, id platform.DatabaseID) (*cnpgv1.Cluster, error) {
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

	return toCluster(list.Items[0])
}

func toCluster(item unstructured.Unstructured) (*cnpgv1.Cluster, error) {
	var cluster cnpgv1.Cluster
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &cluster)
	return &cluster, err
}

func toDatabase(cluster cnpgv1.Cluster, environmentID platform.EnvironmentID) platform.Database {
	limits := cluster.Spec.Resources.Limits

	database := platform.Database{
		ID:           databaseID(cluster, environmentID),
		Name:         cluster.Labels[databaseLabel],
		Instances:    cluster.Spec.Instances,
		Status:       databaseStatus(cluster),
		StatusReason: cluster.Status.PhaseReason,
		Resources: platform.Resources{
			CPU:    limits[core.ResourceCPU],
			Memory: limits[core.ResourceMemory],
		},
		CreatedAt:  cluster.GetCreationTimestamp().Time,
		PublicHost: cluster.Annotations[annotationDatabaseHost],
	}

	// CNPG image: "ghcr.io/cloudnative-pg/postgresql:17-standard-bookworm"
	if i := strings.LastIndex(cluster.Spec.ImageName, ":"); i != -1 {
		tag := cluster.Spec.ImageName[i+1:]
		database.Version = strings.SplitN(tag, "-", 2)[0]
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

func (c *Client) DatabaseCredentials(ctx context.Context, id platform.DatabaseID) (*platform.DatabaseCredentials, error) {
	secretName := "lucity-app-pg-" + id.Name + "-app"
	namespace := id.Namespace()

	secret, err := c.kubernetes.CoreV1().Secrets(namespace).Get(ctx, secretName, meta.GetOptions{})

	if apierrors.IsNotFound(err) {
		return nil, platform.ErrDatabaseProvisioning
	}

	if err != nil {
		return nil, fmt.Errorf("get cnpg secret %q in %q: %w", secretName, namespace, err)
	}

	host := string(secret.Data["host"])

	// CNPG stores the short service name; qualify with namespace for
	// cross-namespace DNS resolution.
	if !strings.Contains(host, ".") {
		host = host + "." + namespace + ".svc.cluster.local"
	}

	return &platform.DatabaseCredentials{
		Host:     host,
		Port:     string(secret.Data["port"]),
		DBName:   string(secret.Data["dbname"]),
		User:     string(secret.Data["user"]),
		Password: string(secret.Data["password"]),
		URI:      string(secret.Data["fqdn-uri"]),
	}, nil
}

var databaseFailurePhases = map[string]bool{
	cnpgv1.PhaseUnrecoverable:              true,
	cnpgv1.PhaseCannotCreateClusterObjects: true,
	cnpgv1.PhaseFailurePlugin:              true,
	cnpgv1.PhaseUnknownPlugin:              true,
	cnpgv1.PhaseImageCatalogError:          true,
	cnpgv1.PhaseArchitectureBinaryMissing:  true,
}

func databaseStatus(cluster cnpgv1.Cluster) platform.DatabaseStatus {
	desired := cluster.Spec.Instances

	if desired == 0 {
		return platform.DatabaseStopped
	}

	if databaseFailurePhases[cluster.Status.Phase] {
		return platform.DatabaseFailed
	}

	ready := cluster.Status.ReadyInstances

	if ready == 0 {
		if cluster.Status.CurrentPrimary == "" {
			return platform.DatabasePending
		}

		return platform.DatabaseUpdating
	}

	if ready >= desired && cluster.Status.Phase == cnpgv1.PhaseHealthy {
		return platform.DatabaseHealthy
	}

	if ready < desired && cluster.Status.Phase == cnpgv1.PhaseHealthy {
		return platform.DatabaseDegraded
	}

	return platform.DatabaseUpdating
}
