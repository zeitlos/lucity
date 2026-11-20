package cockpit

import (
	"path/filepath"
	"sync"
	"time"

	gh "github.com/google/go-github/v92/github"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"

	"github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/victoria"
	"github.com/zeitlos/lucity/services/cockpit/internal/rybbit"
	"github.com/zeitlos/lucity/services/cockpit/internal/signals"
	"github.com/zeitlos/lucity/services/cockpit/internal/victorialogs"
)

const (
	abuseKey     = "abuse"
	abuseNoteKey = "abuseNote"
	suspendedKey = "suspended"
	customerKey  = "stripeCustomerId"
	personalKey  = "personal"
)

type Client struct {
	environment     string
	snapshotPath    string
	githubCachePath string
	identity        *logto.Client
	billing         *billing.Client
	kubernetes      kubernetes.Interface
	objects         metadata.Interface
	metrics         *victoria.Client
	logs            *victorialogs.Client
	github          *gh.Client
	analytics       *rybbit.Client
	stripeAccount   string

	mu            sync.Mutex
	accounts      map[int64]*Account
	parents       map[string]string
	chartVersions map[string]string
}

// New builds the client. logs is nil when the logs store couldn't be reached
// while connecting. analytics is nil when the environment has no Rybbit
// API key, and stripeAccount is empty when the Stripe key cannot read its
// account.
func New(environment, cacheDirectory string, identity *logto.Client, billing *billing.Client, kubernetes kubernetes.Interface, objects metadata.Interface, metrics *victoria.Client, logs *victorialogs.Client, github *gh.Client, analytics *rybbit.Client, stripeAccount string) *Client {
	client := &Client{
		environment:     environment,
		snapshotPath:    filepath.Join(cacheDirectory, "workspaces-"+environment+".json"),
		githubCachePath: filepath.Join(cacheDirectory, "github-"+environment+".json"),
		identity:        identity,
		billing:         billing,
		kubernetes:      kubernetes,
		objects:         objects,
		metrics:         metrics,
		logs:            logs,
		github:          github,
		analytics:       analytics,
		stripeAccount:   stripeAccount,
		accounts:        map[int64]*Account{},
		parents:         map[string]string{},
		chartVersions:   map[string]string{},
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
	PaymentMethod   PaymentMethod
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
	CustomerURL      string
	SubscriptionURL  string
}

// PaymentMethod says whether a workspace has a payment method on file, or
// is held back by the card gate until it adds a card.
type PaymentMethod string

const (
	PaymentMethodNone     PaymentMethod = "none"
	PaymentMethodOnFile   PaymentMethod = "on-file"
	PaymentMethodRequired PaymentMethod = "required"
)

type Member struct {
	ID           string
	Name         string
	Email        string
	Avatar       string
	Role         string
	SignedUpAt   time.Time
	LastSignInAt time.Time
	Account      *Account
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
