package buildjob

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/go-containerregistry/pkg/name"
)

type Interface interface {
	Get(ctx context.Context, id BuildID) (*Job, error)
	List(ctx context.Context, workspaceID, repoURL, contextPath string) ([]Job, error)
	Start(ctx context.Context, opts StartOptions) (*Job, error)
	Cancel(ctx context.Context, id BuildID) (*Job, error)
	Logs(ctx context.Context, id BuildID) (io.ReadCloser, error)
}

type StartOptions struct {
	Workspace        string
	RepoURL          string
	Commit           string
	ContextPath      string
	TargetImageNames []string
	Token            string
	BuildVars        map[string]string
}

type Job struct {
	ID          BuildID
	Status      Status
	SourceURL   string
	Commit      string
	ContextPath string
	TriggeredBy string
	StartedAt   *time.Time
	FinishedAt  *time.Time
	ImageRefs   map[string]name.Reference
}

func (j *Job) ImageRef(imageName string) (name.Reference, error) {
	ref, ok := j.ImageRefs[imageName]

	if !ok {
		return nil, fmt.Errorf("image ref for %q not defined on job %q", imageName, j.ID)
	}

	return ref, nil
}

type Status string

const (
	StatusQueued     Status = "queued"
	StatusRunning    Status = "running"
	StatusSucceeded  Status = "succeeded"
	StatusFailed     Status = "failed"
	StatusCancelling Status = "cancelling"
	StatusCancelled  Status = "cancelled"
)

type BuildID struct {
	Workspace string
	Name      string
}

func ParseBuildID(s string) (BuildID, error) {
	workspace, name, ok := strings.Cut(s, "/")

	if !ok || workspace == "" || name == "" {
		return BuildID{}, fmt.Errorf("invalid build id %q", s)
	}

	return BuildID{Workspace: workspace, Name: name}, nil
}

func (b BuildID) String() string {
	return b.Workspace + "/" + b.Name
}

func (b BuildID) WorkspaceID() string {
	return b.Workspace
}

func (b *BuildID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return fmt.Errorf("BuildID must be a string")
	}

	parsed, err := ParseBuildID(str)

	if err != nil {
		return err
	}

	*b = parsed

	return nil
}

func (b BuildID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(b.String()).MarshalGQL(w)
}

var (
	_ graphql.Marshaler   = BuildID{}
	_ graphql.Unmarshaler = (*BuildID)(nil)
)
