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

type KeyValueStore struct {
	ID        KeyValueStoreID
	Name      string
	Version   string
	Status    DatabaseStatus
	Size      resource.Quantity
	CreatedAt time.Time
}

type KeyValueStoreCredentials struct {
	Host     string
	Port     string
	Password string
	URI      string
}

type KeyValueStoreID struct {
	Workspace   string
	Project     string
	Environment string
	Name        string
}

func ParseKeyValueStoreID(s string) (KeyValueStoreID, error) {
	parts := strings.SplitN(s, "/", 4)

	if len(parts) != 4 || slices.Contains(parts, "") {
		return KeyValueStoreID{}, fmt.Errorf("invalid key-value store id %q", s)
	}

	return KeyValueStoreID{
		Workspace: parts[0], Project: parts[1],
		Environment: parts[2], Name: parts[3],
	}, nil
}

func (k KeyValueStoreID) String() string {
	return k.Workspace + "/" + k.Project + "/" + k.Environment + "/" + k.Name
}

func (k KeyValueStoreID) EnvironmentID() EnvironmentID {
	return EnvironmentID{Workspace: k.Workspace, Project: k.Project, Name: k.Environment}
}

func (k KeyValueStoreID) Namespace() string {
	return k.EnvironmentID().Namespace()
}

func (k KeyValueStoreID) WorkspaceID() string {
	return k.Workspace
}

func (k *KeyValueStoreID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return fmt.Errorf("KeyValueStoreID must be a string")
	}

	parsed, err := ParseKeyValueStoreID(str)

	if err != nil {
		return err
	}

	*k = parsed

	return nil
}

func (k KeyValueStoreID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(k.String()).MarshalGQL(w)
}

var (
	_ WorkspaceScoped     = (*KeyValueStoreID)(nil)
	_ graphql.Marshaler   = KeyValueStoreID{}
	_ graphql.Unmarshaler = (*KeyValueStoreID)(nil)
)
