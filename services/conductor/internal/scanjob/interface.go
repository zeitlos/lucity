package scanjob

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
	Get(ctx context.Context, id ScanID) (*Job, error)
	List(ctx context.Context, service platform.ServiceID) ([]Job, error)
	Start(ctx context.Context, opts StartOptions) (*Job, error)
	Logs(ctx context.Context, id ScanID) (io.ReadCloser, error)
}

type StartOptions struct {
	Service   platform.ServiceID
	Scanner   string
	BuildName string
	SourceURL string
	Commit    string
	Token     string
	ReleaseID string
}

type Job struct {
	ID            ScanID
	Scanner       string
	Status        Status
	Service       platform.ServiceID
	ReleaseID     string
	BuildName     string
	FindingsCount *int
	CreatedAt     time.Time
	StartedAt     *time.Time
	FinishedAt    *time.Time
}

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

var Scanners = []string{"gitleaks", "trufflehog"}

type ScanID struct {
	Workspace string
	Name      string
}

func ParseScanID(s string) (ScanID, error) {
	workspace, name, ok := strings.Cut(s, "/")

	if !ok || workspace == "" || name == "" {
		return ScanID{}, fmt.Errorf("invalid scan id %q", s)
	}

	return ScanID{Workspace: workspace, Name: name}, nil
}

func (s ScanID) String() string {
	return s.Workspace + "/" + s.Name
}

func (s ScanID) WorkspaceID() string {
	return s.Workspace
}

func (s *ScanID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return fmt.Errorf("ScanID must be a string")
	}

	parsed, err := ParseScanID(str)

	if err != nil {
		return err
	}

	*s = parsed

	return nil
}

func (s ScanID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(s.String()).MarshalGQL(w)
}

var (
	_ platform.WorkspaceScoped = ScanID{}
	_ graphql.Marshaler        = ScanID{}
	_ graphql.Unmarshaler      = (*ScanID)(nil)
)
