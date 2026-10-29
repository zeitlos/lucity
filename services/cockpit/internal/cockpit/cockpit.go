package cockpit

import (
	"path/filepath"
	"sync"
	"time"

	gh "github.com/google/go-github/v92/github"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/victoria"
	"github.com/zeitlos/lucity/services/cockpit/internal/signals"
)

const (
	abuseKey     = "abuse"
	abuseNoteKey = "abuseNote"
	suspendedKey = "suspended"
	customerKey  = "stripeCustomerId"
)

type Client struct {
	environment     string
	snapshotPath    string
	githubCachePath string
	identity        *logto.Client
	billing         *billing.Client
	kubernetes      kubernetes.Interface
	metrics         *victoria.Client
	github          *gh.Client

	mu       sync.Mutex
	accounts map[int64]*Account
	parents  map[string]string
}

func New(environment, cacheDirectory string, identity *logto.Client, billing *billing.Client, kubernetes kubernetes.Interface, metrics *victoria.Client, github *gh.Client) *Client {
	client := &Client{
		environment:     environment,
		snapshotPath:    filepath.Join(cacheDirectory, "workspaces-"+environment+".json"),
		githubCachePath: filepath.Join(cacheDirectory, "github-"+environment+".json"),
		identity:        identity,
		billing:         billing,
		kubernetes:      kubernetes,
		metrics:         metrics,
		github:          github,
		accounts:        map[int64]*Account{},
		parents:         map[string]string{},
	}

	client.loadGitHubCache()

	return client
}

func (c *Client) Environment() string {
	return c.environment
}

type Workspace struct {
	ID              string
	CreatedAt       time.Time
	Abuse           bool
	AbuseNote       string
	Suspended       bool
	Billing         *Billing
	Members         []Member
	Signals         []signals.Signal
	RunningPods     int
	PendingPods     int
	MemoryRequested int64
	Egress7d        int64
	Environments    []Environment
}

type Billing struct {
	Plan             billing.Plan
	TrialEndsAt      time.Time
	GoodStanding     bool
	StandingReason   string
	HasPaymentMethod bool
}

type Member struct {
	Name    string
	Email   string
	Role    string
	Account *Account
}

type Account struct {
	Login     string
	CreatedAt time.Time
}

type Environment struct {
	ID              platform.EnvironmentID
	RunningPods     int
	PendingPods     int
	MemoryRequested int64
	Services        []Service
}

type Service struct {
	Name       string
	Image      string
	Commit     string
	Repository string
}

type Result struct {
	WorkspaceID string
	Err         error
}
