package scanjob

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"

	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/usererror"
)

type Interface interface {
	Get(ctx context.Context, id ScanID) (*Job, error)
	List(ctx context.Context, service platform.ServiceID) ([]Job, error)
	Start(ctx context.Context, opts StartOptions) (*Job, error)
	Logs(ctx context.Context, id ScanID) (io.ReadCloser, error)
}

type StartOptions struct {
	Service   platform.ServiceID
	BuildName string
	SourceURL string
	Commit    string
	Token     string
	ReleaseID string
}

type Job struct {
	ID            ScanID
	Status        Status
	Service       platform.ServiceID
	ReleaseID     string
	BuildName     string
	FindingsCount *int
	VerifiedCount *int
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

type ScanID struct {
	Workspace string
	Name      string
}

func ParseScanID(s string) (ScanID, error) {
	workspace, name, ok := strings.Cut(s, "/")

	if !ok || workspace == "" || name == "" {
		return ScanID{}, usererror.Errorf("invalid scan id %q", s)
	}

	return ScanID{Workspace: workspace, Name: name}, nil
}

func (s ScanID) LogValue() slog.Value {
	return slog.StringValue(s.String())
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
		return usererror.New("ScanID must be a string")
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
