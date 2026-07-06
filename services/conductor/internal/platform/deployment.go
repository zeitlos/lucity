package platform

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"

	"github.com/zeitlos/lucity/services/conductor/internal/image"
)

type Deployment struct {
	ID DeploymentID

	Image                image.Ref
	Commit               string
	CommitMessage        string
	Ref                  string
	GitHubInstallationID int

	SourceURL   string
	ContextPath string
	Resources   Resources
	Command     string

	Status   DeploymentStatus
	Replicas ReplicaCount
	Rollout  *Rollout

	BuildID        string
	ReleaseID      string
	ReleaseTrigger string
	ReleaseActor   string
	CreatedAt      time.Time
}

type DeploymentStatus string

const (
	DeploymentDeploying  DeploymentStatus = "deploying"
	DeploymentActive     DeploymentStatus = "active"
	DeploymentSuperseded DeploymentStatus = "superseded"
	DeploymentFailed     DeploymentStatus = "failed"
)

type Rollout struct {
	Status    RolloutStatus
	Reason    RolloutReason
	Message   string
	Restarts  int
	StartedAt time.Time
}

type RolloutStatus string

const (
	RolloutProgressing RolloutStatus = "progressing"
	RolloutReady       RolloutStatus = "ready"
	RolloutDegraded    RolloutStatus = "degraded"
	RolloutFailed      RolloutStatus = "failed"
	RolloutSuperseded  RolloutStatus = "superseded"
)

type RolloutReason string

const (
	RolloutReasonNone             RolloutReason = ""
	RolloutReasonCrashLoop        RolloutReason = "crash_loop"
	RolloutReasonOOMKilled        RolloutReason = "oom_killed"
	RolloutReasonImagePullFailed  RolloutReason = "image_pull_failed"
	RolloutReasonConfigError      RolloutReason = "config_error"
	RolloutReasonQuotaExceeded    RolloutReason = "quota_exceeded"
	RolloutReasonUnschedulable    RolloutReason = "unschedulable"
	RolloutReasonNotReady         RolloutReason = "not_ready"
	RolloutReasonDeadlineExceeded RolloutReason = "deadline_exceeded"
)

type DeploymentID struct {
	Workspace   string
	Project     string
	Environment string
	Service     string
	Hash        string
}

func ParseDeploymentID(s string) (DeploymentID, error) {
	parts := strings.SplitN(s, "/", 5)

	if len(parts) != 5 || slices.Contains(parts, "") {
		return DeploymentID{}, fmt.Errorf("invalid deployment id %q", s)
	}

	return DeploymentID{
		Workspace:   parts[0],
		Project:     parts[1],
		Environment: parts[2],
		Service:     parts[3],
		Hash:        parts[4],
	}, nil
}

func (d DeploymentID) String() string {
	return d.Workspace + "/" + d.Project + "/" + d.Environment + "/" + d.Service + "/" + d.Hash
}

func (d DeploymentID) ServiceID() ServiceID {
	return ServiceID{
		Workspace:   d.Workspace,
		Project:     d.Project,
		Environment: d.Environment,
		Name:        d.Service,
	}
}

func (d DeploymentID) Namespace() string {
	return d.ServiceID().Namespace()
}

func (d DeploymentID) WorkspaceID() string {
	return d.Workspace
}

func (d *DeploymentID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return fmt.Errorf("DeploymentID must be a string")
	}

	parsed, err := ParseDeploymentID(str)

	if err != nil {
		return err
	}

	*d = parsed

	return nil
}

func (d DeploymentID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(d.String()).MarshalGQL(w)
}

var (
	_ WorkspaceScoped     = (*DeploymentID)(nil)
	_ graphql.Marshaler   = DeploymentID{}
	_ graphql.Unmarshaler = (*DeploymentID)(nil)
)
