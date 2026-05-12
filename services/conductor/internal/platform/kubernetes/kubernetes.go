package kubernetes

import (
	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	apps "k8s.io/api/apps/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const (
	workspaceLabel       = "lucity.dev/workspace"
	projectLabel         = "lucity.dev/project"
	environmentLabel     = "lucity.dev/environment"
	serviceLabel         = "lucity.dev/service"
	volumeLabel          = "lucity.dev/volume"
	databaseLabel        = "lucity.dev/database"
	podTemplateHashLabel = apps.DefaultDeploymentUniqueLabelKey
	resourceTierLabel    = "lucity.dev/resource-tier"
)

const (
	resourceTierEco  = "eco"
	resourceTierProd = "production"
)

type Client struct {
	kubernetes kubernetes.Interface
	dynamic    dynamic.Interface
}

func New(kubernetes kubernetes.Interface, dynamic dynamic.Interface) (*Client, error) {
	return &Client{}, nil
}

var _ platform.Interface = (*Client)(nil)
