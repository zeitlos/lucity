package kubernetes

import (
	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	apps "k8s.io/api/apps/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const (
	workspaceLabel          = "lucity.dev/workspace"
	projectLabel            = "lucity.dev/project"
	environmentLabel        = "lucity.dev/environment"
	serviceLabel            = "lucity.dev/service"
	volumeLabel             = "lucity.dev/volume"
	databaseLabel           = "lucity.dev/database"
	keyValueStoreLabel      = "lucity.dev/keyvaluestore"
	podTemplateHashLabel    = apps.DefaultDeploymentUniqueLabelKey
	resourceTierLabel       = "lucity.dev/resource-tier"
	gitHubInstallationLabel = "lucity.dev/github-installation"
)

const (
	resourceTierEco  = "eco"
	resourceTierProd = "production"
)

const (
	annotationSourceRepo     = "lucity.dev/source-repo"
	annotationSourceCommit   = "lucity.dev/source-commit"
	annotationSourceMessage  = "lucity.dev/source-commit-message"
	annotationSourceRef      = "lucity.dev/source-ref"
	annotationSourceContext  = "lucity.dev/source-context"
	annotationSourceBranch   = "lucity.dev/source-branch"
	annotationAutoDeploy     = "lucity.dev/autodeploy"
	annotationBuildID        = "lucity.dev/build-id"
	annotationRelease        = "lucity.dev/release"
	annotationReleaseTrigger = "lucity.dev/release-trigger"
	annotationReleaseActor   = "lucity.dev/release-actor"
	annotationRevision       = "deployment.kubernetes.io/revision"
	annotationDatabaseHost   = "lucity.dev/db-host"
)

type Client struct {
	kubernetes      kubernetes.Interface
	dynamic         dynamic.Interface
	variableSources []variableSource
}

func New(kubernetes kubernetes.Interface, dynamic dynamic.Interface) *Client {
	client := &Client{
		kubernetes: kubernetes,
		dynamic:    dynamic,
	}

	client.variableSources = defaultVariableSources(kubernetes)

	return client
}

var _ platform.Interface = (*Client)(nil)
