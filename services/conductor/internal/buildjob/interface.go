package buildjob

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
)

type Interface interface {
	Get(ctx context.Context, id string) (*Job, error)
	List(ctx context.Context, workspaceID, repoURL, contextPath string) ([]Job, error)
	Start(ctx context.Context, opts StartOptions) (*Job, error)
	Cancel(ctx context.Context, id string) (*Job, error)
	Logs(ctx context.Context, id string) (io.ReadCloser, error)
}

type StartOptions struct {
	Workspace        string
	RepoURL          string
	Commit           string
	ContextPath      string
	TargetImageNames []string
	Token            string
}

type Job struct {
	ID          string
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
