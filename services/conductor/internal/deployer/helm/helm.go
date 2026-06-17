package helm

import (
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
)

type Client struct {
	chart            *chart.Chart
	settings         *cli.EnvSettings
	gatewayName      string
	gatewayNamespace string

	defaultCPULimit    resource.Quantity
	defaultMemoryLimit resource.Quantity
}

func New(chart *chart.Chart, gatewayName, gatewayNamespace string, defaultCPULimit, defaultMemoryLimit resource.Quantity) *Client {
	return &Client{
		chart:            chart,
		settings:         cli.New(),
		gatewayName:      gatewayName,
		gatewayNamespace: gatewayNamespace,

		defaultCPULimit:    defaultCPULimit,
		defaultMemoryLimit: defaultMemoryLimit,
	}
}

func (c *Client) Services() deployer.ServiceClient {
	return &serviceClient{client: c}
}

func (c *Client) Databases() deployer.DatabaseClient {
	return &databaseClient{client: c}
}

func (c *Client) Volumes() deployer.VolumeClient {
	return &volumeClient{client: c}
}

func (c *Client) Environments() deployer.EnvironmentClient {
	return &environmentClient{client: c}
}

var _ deployer.Interface = (*Client)(nil)
