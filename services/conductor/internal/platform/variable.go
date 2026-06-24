package platform

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/99designs/gqlgen/graphql"
)

type Variable struct {
	ID     VariableID
	Name   string
	Source VariableSource
}

type VariableSource struct {
	Database      *DatabaseID
	KeyValueStore *KeyValueStoreID
	Bucket        *BucketID
	Shared        bool
}

type VariableID struct {
	Workspace   string
	Project     string
	Environment string
	Secret      string
	Name        string
}

func ParseVariableID(s string) (VariableID, error) {
	parts := strings.SplitN(s, "/", 5)

	if len(parts) != 5 || slices.Contains(parts, "") {
		return VariableID{}, fmt.Errorf("invalid variable id %q", s)
	}

	return VariableID{
		Workspace:   parts[0],
		Project:     parts[1],
		Environment: parts[2],
		Secret:      parts[3],
		Name:        parts[4],
	}, nil
}

func (v VariableID) String() string {
	return v.Workspace + "/" + v.Project + "/" + v.Environment + "/" + v.Secret + "/" + v.Name
}

func (v VariableID) EnvironmentID() EnvironmentID {
	return EnvironmentID{Workspace: v.Workspace, Project: v.Project, Name: v.Environment}
}

func (v VariableID) WorkspaceID() string {
	return v.Workspace
}

func (v *VariableID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return fmt.Errorf("VariableID must be a string")
	}

	parsed, err := ParseVariableID(str)

	if err != nil {
		return err
	}

	*v = parsed

	return nil
}

func (v VariableID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(v.String()).MarshalGQL(w)
}

var (
	_ WorkspaceScoped     = (*VariableID)(nil)
	_ graphql.Marshaler   = VariableID{}
	_ graphql.Unmarshaler = (*VariableID)(nil)
)
