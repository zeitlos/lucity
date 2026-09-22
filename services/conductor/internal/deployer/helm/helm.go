package helm

import (
	"github.com/blang/semver/v4"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/values"

	"helm.sh/helm/v3/pkg/chart"
)

type Client struct {
	chartVersion     semver.Version
	chart            *chart.Chart
	gatewayName      string
	gatewayNamespace string
	headerMatches    []values.HeaderMatch
	clusterIssuer    string
	backups          BackupConfig
}

type Option func(*Client)

func WithHeaderMatches(matches ...values.HeaderMatch) Option {
	return func(c *Client) {
		c.headerMatches = matches
	}
}

type BackupConfig struct {
	Enabled  bool
	Endpoint string
	Bucket   string
}

func New(chart *chart.Chart, gatewayName, gatewayNamespace, clusterIssuer string, backups BackupConfig, options ...Option) (*Client, error) {
	chartVersion, err := semver.Parse(chart.Metadata.Version)

	if err != nil {
		return nil, err
	}

	client := &Client{
		chartVersion:     chartVersion,
		chart:            chart,
		gatewayName:      gatewayName,
		gatewayNamespace: gatewayNamespace,
		clusterIssuer:    clusterIssuer,
		backups:          backups,
	}

	for _, option := range options {
		option(client)
	}

	return client, nil
}

func (c *Client) Services() deployer.ServiceClient {
	return &serviceClient{client: c}
}

func (c *Client) Databases() deployer.DatabaseClient {
	return &databaseClient{client: c}
}

func (c *Client) KeyValueStores() deployer.KeyValueStoreClient {
	return &keyValueStoreClient{client: c}
}

func (c *Client) Volumes() deployer.VolumeClient {
	return &volumeClient{client: c}
}

func (c *Client) Environments() deployer.EnvironmentClient {
	return &environmentClient{client: c}
}

var _ deployer.Interface = (*Client)(nil)
