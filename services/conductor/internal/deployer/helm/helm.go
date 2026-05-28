package helm

import (
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
)

type Client struct {
	chart      *chart.Chart
	restGetter genericclioptions.RESTClientGetter
	settings   *cli.EnvSettings
}

func New(chart *chart.Chart, restGetter genericclioptions.RESTClientGetter) *Client {
	return &Client{
		chart:      chart,
		restGetter: restGetter,
		settings:   cli.New(),
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
