package kubernetes

import (
	"context"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/environment"
	"github.com/zeitlos/lucity/pkg/platform"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var _ environment.Interface = (*Client)(nil)

const (
	resourceQuotaName        = "lucity-resources"
	limitRangeName           = "lucity-defaults"
	PullSecretName           = deployer.PullSecretName
	BackupSecretName         = deployer.BackupSecretName
	BackupAccessKeyIDKey     = deployer.BackupAccessKeyIDKey
	BackupSecretAccessKeyKey = deployer.BackupSecretAccessKeyKey
	BackupRegionKey          = deployer.BackupRegionKey

	pullSecretMaxAge = time.Minute
)

type Client struct {
	k8s              kubernetes.Interface
	dyn              dynamic.Interface
	systemNamespace  string
	systemPullSecret string
	podCIDR          string
	serviceCIDR      string
	backups          BackupCredentials

	mu             sync.Mutex
	pullSecret     *corev1.Secret
	pullSecretRead time.Time
}

type BackupCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

func New(k8s kubernetes.Interface, dyn dynamic.Interface, systemNamespace, systemPullSecret, podCIDR, serviceCIDR string, backups BackupCredentials) *Client {
	return &Client{
		k8s:              k8s,
		dyn:              dyn,
		systemNamespace:  systemNamespace,
		systemPullSecret: systemPullSecret,
		podCIDR:          podCIDR,
		serviceCIDR:      serviceCIDR,
		backups:          backups,
	}
}

// Ensure runs the full per-env scaffolding. Order matters: the namespace
// must exist before everything else, and the NetworkPolicies must land
// before quotas/limit-ranges/pull-secrets so the wall is up the moment any
// workload could be scheduled.
func (c *Client) Ensure(ctx context.Context, id platform.EnvironmentID, tier platform.ResourceTier) error {
	if err := c.ensureNamespace(ctx, id, tier); err != nil {
		return err
	}

	if err := c.ensureNetworkPolicy(ctx, id); err != nil {
		return err
	}

	if err := c.ensureCiliumNetworkPolicy(ctx, id); err != nil {
		return err
	}

	if err := c.ensureQuota(ctx, id); err != nil {
		return err
	}

	if err := c.ensureLimitRange(ctx, id, tier); err != nil {
		return err
	}

	if err := c.ensurePullSecret(ctx, id); err != nil {
		return err
	}

	if err := c.ensureBackupSecret(ctx, id); err != nil {
		return err
	}

	return nil
}

func (c *Client) Delete(ctx context.Context, id platform.EnvironmentID) error {
	return c.deleteNamespace(ctx, id)
}
