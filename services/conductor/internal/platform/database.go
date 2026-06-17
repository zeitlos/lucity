package platform

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"k8s.io/apimachinery/pkg/api/resource"
)

// TODO: Rename to Postgres, PostgresDatabase or PostgresCluster
type Database struct {
	ID         DatabaseID
	Name       string
	Version    string
	Instances  int
	Status     DatabaseStatus
	Size       resource.Quantity
	CreatedAt  time.Time
	PublicHost string
}

type DatabaseStatus string

const (
	DatabaseHealthy  DatabaseStatus = "healthy"
	DatabaseDegraded DatabaseStatus = "degraded"
	DatabaseFailed   DatabaseStatus = "failed"
	DatabasePending  DatabaseStatus = "pending"
	DatabaseStopped  DatabaseStatus = "stopped"
)

type DatabaseID struct {
	Workspace   string
	Project     string
	Environment string
	Name        string
}

func ParseDatabaseID(s string) (DatabaseID, error) {
	parts := strings.SplitN(s, "/", 4)

	if len(parts) != 4 || slices.Contains(parts, "") {
		return DatabaseID{}, fmt.Errorf("invalid database id %q", s)
	}

	return DatabaseID{
		Workspace: parts[0], Project: parts[1],
		Environment: parts[2], Name: parts[3],
	}, nil
}

func (d DatabaseID) String() string {
	return d.Workspace + "/" + d.Project + "/" + d.Environment + "/" + d.Name
}

func (d DatabaseID) EnvironmentID() EnvironmentID {
	return EnvironmentID{Workspace: d.Workspace, Project: d.Project, Name: d.Environment}
}

func (d DatabaseID) Namespace() string {
	return d.EnvironmentID().Namespace()
}

func (d DatabaseID) WorkspaceID() string {
	return d.Workspace
}

func (d *DatabaseID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return fmt.Errorf("DatabaseID must be a string")
	}

	parsed, err := ParseDatabaseID(str)

	if err != nil {
		return err
	}

	*d = parsed

	return nil
}

func (d DatabaseID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(d.String()).MarshalGQL(w)
}

var (
	_ WorkspaceScoped     = (*DatabaseID)(nil)
	_ graphql.Marshaler   = DatabaseID{}
	_ graphql.Unmarshaler = (*DatabaseID)(nil)
)
