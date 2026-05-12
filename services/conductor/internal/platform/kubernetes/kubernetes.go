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

const (
	annotationSourceRepo   = "lucity.dev/source-repo"
	annotationSourceCommit = "lucity.dev/source-commit"
	annotationSourceRef    = "lucity.dev/source-ref"
	annotationImageDigest  = "lucity.dev/image-digest"
)

type Client struct {
	kubernetes kubernetes.Interface
	dynamic    dynamic.Interface
}

func New(kubernetes kubernetes.Interface, dynamic dynamic.Interface) *Client {
	return &Client{
		kubernetes: kubernetes,
		dynamic:    dynamic,
	}
}

var _ platform.Interface = (*Client)(nil)
