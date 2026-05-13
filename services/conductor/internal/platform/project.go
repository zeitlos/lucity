package platform

import (
	"fmt"
	"io"
	"strings"

	"github.com/99designs/gqlgen/graphql"
)

type Project struct {
	ID           ProjectID
	Name         string
	Environments []Environment
}

type ProjectID struct {
	Workspace string
	Name      string
}

func ParseProjectID(s string) (ProjectID, error) {
	ws, name, ok := strings.Cut(s, "/")

	if !ok || ws == "" || name == "" {
		return ProjectID{}, fmt.Errorf("invalid project id %q", s)
	}

	return ProjectID{Workspace: ws, Name: name}, nil
}

func (p ProjectID) String() string {
	return p.Workspace + "/" + p.Name
}

func (p ProjectID) WorkspaceID() string {
	return p.Workspace
}

func (p *ProjectID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return fmt.Errorf("ProjectID must be a string")
	}

	parsed, err := ParseProjectID(str)

	if err != nil {
		return err
	}

	*p = parsed

	return nil
}

func (p ProjectID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(p.String()).MarshalGQL(w)
}

var (
	_ WorkspaceScoped     = (*ProjectID)(nil)
	_ graphql.Marshaler   = ProjectID{}
	_ graphql.Unmarshaler = (*ProjectID)(nil)
)
