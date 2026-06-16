package platform

import (
	"crypto/sha256"
	"encoding/hex"
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
	Variables    map[string]string // shared variables; nil if release not yet installed
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

const namespaceHashLen = 10

func (e EnvironmentID) Namespace() string {
	// Suffixing with hash of canonical env id to avoid namespace collisions which could result in a takeover.
	sum := sha256.Sum256([]byte(e.String()))
	return e.Workspace + "-" + e.Project + "-" + e.Name + "-" + hex.EncodeToString(sum[:])[:namespaceHashLen]
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
