package helm

import (
	"github.com/blang/semver/v4"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/deployer/values"

	chart "helm.sh/helm/v4/pkg/chart/v2"
)

type Client struct {
	chartVersion         semver.Version
	chart                *chart.Chart
	gatewayName          string
	gatewayNamespace     string
	gatewayHTTPSListener string
	headerMatches        []values.HeaderMatch
	clusterIssuer        string
	backups              BackupConfig
	registry             string
	defaultPlacement     values.Placement
	agents               AgentConfig
	releases             *ReleaseCache
}

type Option func(*Client)

func WithHTTPSListener(name string) Option {
	return func(c *Client) {
		c.gatewayHTTPSListener = name
	}
}

func WithHeaderMatches(matches ...values.HeaderMatch) Option {
	return func(c *Client) {
		c.headerMatches = matches
	}
}

// WithEdgeHeader makes workload routes match only requests that carry the
// edge header, when enforced is true.
func WithEdgeHeader(enforced bool) Option {
	return func(c *Client) {
		if !enforced {
			return
		}

		c.headerMatches = []values.HeaderMatch{{
			Name:      deployer.EdgeHeaderName,
			ValueFrom: &values.SecretKeyRef{SecretName: deployer.EdgeHeaderSecretName, Key: deployer.EdgeHeaderSecretKey},
		}}
	}
}

func WithReleaseCache(releases *ReleaseCache) Option {
	return func(c *Client) {
		c.releases = releases
	}
}

type BackupConfig struct {
	Enabled  bool
	Endpoint string
	Bucket   string
}

// AgentConfig is what agents need from the platform: the runner image, the
// agent gateway's base URL (without a workspace) and namespace, and the chart
// of the workspace releases that live next to the gateway. A zero AgentConfig
// leaves the agents of a release as they are and refuses workspace changes.
type AgentConfig struct {
	Image          string
	GatewayURL     string
	Namespace      string
	WorkspaceChart *chart.Chart
}

// The shared agent gateway, and the listener workspace routes attach to.
const (
	AgentGatewayName     = "lucity-agent-gateway"
	AgentGatewayListener = "http"
)

func New(chart *chart.Chart, gatewayName, gatewayNamespace, clusterIssuer string, backups BackupConfig, registry string, defaultPlacement values.Placement, agents AgentConfig, options ...Option) (*Client, error) {
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
		registry:         registry,
		defaultPlacement: defaultPlacement,
		agents:           agents,
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

func (c *Client) Agents() deployer.AgentClient {
	return &agentClient{client: c}
}

func (c *Client) Workspaces() deployer.WorkspaceClient {
	return &workspaceClient{client: c}
}

var _ deployer.Interface = (*Client)(nil)
