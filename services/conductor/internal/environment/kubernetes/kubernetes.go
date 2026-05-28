package kubernetes

import (
	"context"

	"github.com/zeitlos/lucity/services/conductor/internal/environment"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	"k8s.io/client-go/kubernetes"
)

var _ environment.Interface = (*Client)(nil)

const (
	resourceQuotaName = "lucity-resources"
	limitRangeName    = "lucity-defaults"
	pullSecretName    = "lucity-registry"
)

type Client struct {
	k8s              kubernetes.Interface
	systemNamespace  string
	systemPullSecret string
}

func New(k8s kubernetes.Interface, systemNamespace, systemPullSecret string) *Client {
	return &Client{
		k8s:              k8s,
		systemNamespace:  systemNamespace,
		systemPullSecret: systemPullSecret,
	}
}

func (c *Client) Ensure(ctx context.Context, id platform.EnvironmentID, tier platform.ResourceTier) error {
	if err := c.ensureNamespace(ctx, id, tier); err != nil {
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

	return nil
}

func (c *Client) Delete(ctx context.Context, id platform.EnvironmentID) error {
	return c.deleteNamespace(ctx, id)
}
