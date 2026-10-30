package conductor

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/domains"
	"github.com/zeitlos/lucity/pkg/environment"
	ghpkg "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/hostname"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/pkg/pipeline"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/usererror"
	"github.com/zeitlos/lucity/services/conductor/internal/agentaudit"
	"github.com/zeitlos/lucity/services/conductor/internal/agentrunner"
	"github.com/zeitlos/lucity/services/conductor/internal/backuparchive"
	"github.com/zeitlos/lucity/services/conductor/internal/billing"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	"github.com/zeitlos/lucity/services/conductor/internal/deployjob"
	"github.com/zeitlos/lucity/services/conductor/internal/directory"
	"github.com/zeitlos/lucity/services/conductor/internal/metrics"
	"github.com/zeitlos/lucity/services/conductor/internal/objectstorage"
	"github.com/zeitlos/lucity/services/conductor/internal/planner"
	"github.com/zeitlos/lucity/services/conductor/internal/registry"
	"github.com/zeitlos/lucity/services/conductor/internal/scanjob"
	"github.com/zeitlos/lucity/services/conductor/internal/scanreport"
	"github.com/zeitlos/lucity/services/conductor/internal/source"
	"github.com/zeitlos/lucity/services/conductor/internal/upstream"
	"github.com/zeitlos/lucity/services/conductor/internal/vulnerabilities"
	"k8s.io/apimachinery/pkg/api/resource"
)

type Client struct {
	billing   *billing.Client
	gitHubApp *ghpkg.App
	logto     *logto.Client

	directory       directory.Interface
	platform        platform.Interface
	buildjob        buildjob.Interface
	deployjob       deployjob.Interface
	scanjob         scanjob.Interface
	scanreport      *scanreport.Client
	vulnerabilities *vulnerabilities.Client
	pipeline        pipeline.Interface
	planner         planner.Interface
	source          source.Interface
	hostname        *hostname.Client
	deployer        deployer.Interface
	environment     environment.Interface
	objectStorage   objectstorage.Interface
	metrics         *metrics.Provider
	domains         *domains.Client
	registry        *registry.Client
	upstream        *upstream.Client
	agentRunner     *agentrunner.Client
	agentAudit      *agentaudit.Client
	signups         prometheus.Counter

	config Config

	// Cached Logto org role IDs (looked up by name on first use)
	orgRoleOnce  sync.Once
	adminRoleID  string
	memberRoleID string

	orgIDCache   map[string]string
	orgIDCacheMu sync.RWMutex

	capabilitiesCache map[string]cachedCapabilities
	capabilitiesMu    sync.Mutex
}

type Config struct {
	Version              string
	ChartFS              fs.FS
	RegistryPullSecret   authn.Keychain
	RegistryURL          string
	RegistryPushURL      string
	RegistryPullURL      string
	WorkloadDomain       string
	DatabaseDomain       string
	LoadBalancerHostname string
	LoadBalancerIP       string
	GitHubAppSlug        string
	DashboardURL         string
	MaxQueuedReleases    int
	PaidPlacement        values.Placement
	BackupArchive        *backuparchive.Client
	AgentsEnabled        bool
	CardGateAccountAge   time.Duration
}

func New(billing *billing.Client, githubApp *ghpkg.App, logto *logto.Client, directory directory.Interface, platform platform.Interface, buildjob buildjob.Interface, deployjob deployjob.Interface, scanjob scanjob.Interface, scanreport *scanreport.Client, vulnerabilities *vulnerabilities.Client, pipeline pipeline.Interface, planner planner.Interface, source source.Interface, hostname *hostname.Client, deployer deployer.Interface, environment environment.Interface, objectStorage objectstorage.Interface, metrics *metrics.Provider, domains *domains.Client, upstream *upstream.Client, agentRunner *agentrunner.Client, agentAudit *agentaudit.Client, config Config, registerer prometheus.Registerer) *Client {
	signups := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "lucity_signups_total",
		Help: "Users whose personal workspace was created.",
	})

	registerer.MustRegister(signups)

	return &Client{
		billing:           billing,
		gitHubApp:         githubApp,
		logto:             logto,
		config:            config,
		orgIDCache:        make(map[string]string),
		capabilitiesCache: make(map[string]cachedCapabilities),
		directory:         directory,
		platform:          platform,
		buildjob:          buildjob,
		deployjob:         deployjob,
		scanjob:           scanjob,
		scanreport:        scanreport,
		vulnerabilities:   vulnerabilities,
		pipeline:          pipeline,
		planner:           planner,
		source:            source,
		hostname:          hostname,
		deployer:          deployer,
		environment:       environment,
		objectStorage:     objectStorage,
		metrics:           metrics,
		domains:           domains,
		upstream:          upstream,
		agentRunner:       agentRunner,
		agentAudit:        agentAudit,
		signups:           signups,
		registry: registry.New(registry.Config{
			Endpoint:     config.RegistryPullURL,
			DialEndpoint: config.RegistryPushURL,
			Keychain:     config.RegistryPullSecret,
		}),
	}
}

// orgRoleIDs returns the cached admin and member role IDs, looking them up on first call.
func (c *Client) orgRoleIDs(ctx context.Context) (adminID, memberID string, err error) {
	c.orgRoleOnce.Do(func() {
		roles, rolesErr := c.logto.OrganizationRoles(ctx)

		if rolesErr != nil {
			err = fmt.Errorf("failed to fetch organization roles: %w", rolesErr)
			return
		}
		for _, r := range roles {
			switch r.Name {
			case "admin":
				c.adminRoleID = r.ID
			case "member":
				c.memberRoleID = r.ID
			}
		}
		if c.adminRoleID == "" || c.memberRoleID == "" {
			err = fmt.Errorf("missing org roles: admin=%q member=%q", c.adminRoleID, c.memberRoleID)
			return
		}
		slog.Info("logto org roles cached", "admin", c.adminRoleID, "member", c.memberRoleID)
	})
	if err != nil {
		// Reset so next call retries
		c.orgRoleOnce = sync.Once{}
		return "", "", err
	}
	return c.adminRoleID, c.memberRoleID, nil
}

func validateResources(cpu, memory string, cpuLimit, memoryLimit QuantityLimit, current *platform.Resources) (deployer.Resources, error) {
	cpuQuantity, err := resource.ParseQuantity(cpu)

	if err != nil {
		return deployer.Resources{}, usererror.Errorf("invalid cpu value %q", cpu)
	}

	memoryQuantity, err := resource.ParseQuantity(memory)

	if err != nil {
		return deployer.Resources{}, usererror.Errorf("invalid memory value %q", memory)
	}

	var currentCPU, currentMemory *resource.Quantity

	if current != nil {
		currentCPU, currentMemory = &current.CPU, &current.Memory
	}

	if err := cpuLimit.check("CPU", cpuQuantity, currentCPU); err != nil {
		return deployer.Resources{}, err
	}

	if err := memoryLimit.check("Memory", memoryQuantity, currentMemory); err != nil {
		return deployer.Resources{}, err
	}

	return deployer.Resources{
		CPU:    cpuQuantity,
		Memory: memoryQuantity,
	}, nil
}
