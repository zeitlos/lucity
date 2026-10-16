package platform

import (
	"io"
	"slices"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/pkg/usererror"
)

type VolumeID struct {
	Workspace   string
	Project     string
	Environment string
	Name        string
}

type VolumeStatus string

const (
	VolumeReady   VolumeStatus = "ready"
	VolumePending VolumeStatus = "pending"
	VolumeFailed  VolumeStatus = "failed"
)

type VolumeMount struct {
	Service ServiceID
	Path    string
}

type Volume struct {
	ID        VolumeID
	Name      string
	Size      resource.Quantity
	Status    VolumeStatus
	Mount     *VolumeMount
	CreatedAt time.Time
}

func ParseVolumeID(s string) (VolumeID, error) {
	parts := strings.SplitN(s, "/", 4)

	if len(parts) != 4 || slices.Contains(parts, "") {
		return VolumeID{}, usererror.Errorf("invalid volume id %q", s)
	}

	return VolumeID{
		Workspace: parts[0], Project: parts[1],
		Environment: parts[2], Name: parts[3],
	}, nil
}

func (v VolumeID) String() string {
	return v.Workspace + "/" + v.Project + "/" + v.Environment + "/" + v.Name
}

func (v VolumeID) EnvironmentID() EnvironmentID {
	return EnvironmentID{Workspace: v.Workspace, Project: v.Project, Name: v.Environment}
}

func (v VolumeID) Namespace() string {
	return v.EnvironmentID().Namespace()
}

func (v VolumeID) WorkspaceID() string {
	return v.Workspace
}

func (v *VolumeID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return usererror.New("VolumeID must be a string")
	}

	parsed, err := ParseVolumeID(str)

	if err != nil {
		return err
	}

	*v = parsed

	return nil
}

func (v VolumeID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(v.String()).MarshalGQL(w)
}

var (
	_ WorkspaceScoped     = VolumeID{}
	_ graphql.Marshaler   = VolumeID{}
	_ graphql.Unmarshaler = (*VolumeID)(nil)
)
