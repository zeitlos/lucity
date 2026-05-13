package platform

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

type Environment struct {
	ID           EnvironmentID
	Name         string
	ResourceTier ResourceTier
	CreatedAt    time.Time
}

type ResourceTier string

const (
	EcoTier        ResourceTier = "eco"
	ProductionTier ResourceTier = "prod"
)

type EnvironmentID struct {
	Workspace string
	Project   string
	Name      string
}

func ParseEnvironmentID(s string) (EnvironmentID, error) {
	ws, rest, ok := strings.Cut(s, "/")

	if !ok || ws == "" {
		return EnvironmentID{}, fmt.Errorf("invalid environment id %q", s)
	}

	proj, name, ok := strings.Cut(rest, "/")

	if !ok || proj == "" || name == "" {
		return EnvironmentID{}, fmt.Errorf("invalid environment id %q", s)
	}

	return EnvironmentID{Workspace: ws, Project: proj, Name: name}, nil
}

func (e EnvironmentID) String() string {
	return e.Workspace + "/" + e.Project + "/" + e.Name
}

func (e EnvironmentID) ProjectID() ProjectID {
	return ProjectID{Workspace: e.Workspace, Name: e.Project}
}

func (e EnvironmentID) Namespace() string {
	return e.Workspace + "-" + e.Project + "-" + e.Name
}

func (e EnvironmentID) WorkspaceID() string {
	return e.Workspace
}

func (e *EnvironmentID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return fmt.Errorf("EnvironmentID must be a string")
	}

	parsed, err := ParseEnvironmentID(str)

	if err != nil {
		return err
	}

	*e = parsed

	return nil
}

func (e EnvironmentID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(e.String()).MarshalGQL(w)
}

var (
	_ WorkspaceScoped     = (*EnvironmentID)(nil)
	_ graphql.Marshaler   = EnvironmentID{}
	_ graphql.Unmarshaler = (*EnvironmentID)(nil)
)
