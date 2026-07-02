package deployjob

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type Interface interface {
	Get(ctx context.Context, id DeployID) (*Job, error)
	List(ctx context.Context, service platform.ServiceID) ([]Job, error)
	Start(ctx context.Context, opts StartOptions) (*Job, error)
	Logs(ctx context.Context, id DeployID) (io.ReadCloser, error)
}

type StartOptions struct {
	Service        platform.ServiceID
	BuildName      string
	CommitMessage  string
	ReleaseID      string
	ReleaseTrigger string
	ReleaseActor   string
}

type Job struct {
	ID         DeployID
	Status     Status
	Service    platform.ServiceID
	ReleaseID  string
	BuildName  string
	StartedAt  *time.Time
	FinishedAt *time.Time
}

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

type DeployID struct {
	Workspace string
	Name      string
}

func ParseDeployID(s string) (DeployID, error) {
	workspace, name, ok := strings.Cut(s, "/")

	if !ok || workspace == "" || name == "" {
		return DeployID{}, fmt.Errorf("invalid deploy id %q", s)
	}

	return DeployID{Workspace: workspace, Name: name}, nil
}

func (d DeployID) String() string {
	return d.Workspace + "/" + d.Name
}

func (d DeployID) WorkspaceID() string {
	return d.Workspace
}

func (d *DeployID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return fmt.Errorf("DeployID must be a string")
	}

	parsed, err := ParseDeployID(str)

	if err != nil {
		return err
	}

	*d = parsed

	return nil
}

func (d DeployID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(d.String()).MarshalGQL(w)
}

var (
	_ platform.WorkspaceScoped = DeployID{}
	_ graphql.Marshaler        = DeployID{}
	_ graphql.Unmarshaler      = (*DeployID)(nil)
)
