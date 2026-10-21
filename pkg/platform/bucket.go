package platform

import (
	"io"
	"log/slog"
	"slices"
	"strings"

	"github.com/99designs/gqlgen/graphql"

	"github.com/zeitlos/lucity/pkg/usererror"
)

type BucketID struct {
	Workspace   string
	Project     string
	Environment string
	Name        string
}

func ParseBucketID(s string) (BucketID, error) {
	parts := strings.SplitN(s, "/", 4)

	if len(parts) != 4 || slices.Contains(parts, "") {
		return BucketID{}, usererror.Errorf("invalid bucket id %q", s)
	}

	return BucketID{
		Workspace: parts[0], Project: parts[1],
		Environment: parts[2], Name: parts[3],
	}, nil
}

func (b BucketID) LogValue() slog.Value {
	return slog.StringValue(b.String())
}

func (b BucketID) String() string {
	return b.Workspace + "/" + b.Project + "/" + b.Environment + "/" + b.Name
}

func (b BucketID) EnvironmentID() EnvironmentID {
	return EnvironmentID{Workspace: b.Workspace, Project: b.Project, Name: b.Environment}
}

func (b BucketID) Namespace() string {
	return b.EnvironmentID().Namespace()
}

func (b BucketID) SecretName() string {
	return "lucity-bucket-" + b.Name
}

func (b BucketID) WorkspaceID() string {
	return b.Workspace
}

func (b *BucketID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return usererror.New("BucketID must be a string")
	}

	parsed, err := ParseBucketID(str)

	if err != nil {
		return err
	}

	*b = parsed

	return nil
}

func (b BucketID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(b.String()).MarshalGQL(w)
}

var (
	_ WorkspaceScoped     = BucketID{}
	_ graphql.Marshaler   = BucketID{}
	_ graphql.Unmarshaler = (*BucketID)(nil)
)
