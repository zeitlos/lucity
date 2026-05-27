package buildjob

import (
	"context"
	"io"
	"time"
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
	ImageRefs   []string
	TriggeredBy string
	StartedAt   *time.Time
	FinishedAt  *time.Time
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
