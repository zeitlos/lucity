package helm

import (
	"github.com/blang/semver/v4"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer"

	"helm.sh/helm/v3/pkg/chart"
	"k8s.io/apimachinery/pkg/api/resource"
)

type Client struct {
	chartVersion     semver.Version
	chart            *chart.Chart
	gatewayName      string
	gatewayNamespace string

	defaultCPULimit    resource.Quantity
	defaultMemoryLimit resource.Quantity
}

func New(chart *chart.Chart, gatewayName, gatewayNamespace string, defaultCPULimit, defaultMemoryLimit resource.Quantity) (*Client, error) {
	chartVersion, err := semver.Parse(chart.Metadata.Version)

	if err != nil {
		return nil, err
	}

	return &Client{
		chartVersion:     chartVersion,
		chart:            chart,
		gatewayName:      gatewayName,
		gatewayNamespace: gatewayNamespace,

		defaultCPULimit:    defaultCPULimit,
		defaultMemoryLimit: defaultMemoryLimit,
	}, nil
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
