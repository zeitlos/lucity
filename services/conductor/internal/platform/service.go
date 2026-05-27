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

type Service struct {
	ID   ServiceID
	Name string

	Status      ServiceStatus
	Replicas    ReplicaCount
	Autoscaling *AutoscalingSettings

	Endpoints []Endpoint

	SourceURL   string
	Branch      string
	ContextPath string
	Resources   Resources
	Command     string

	ActiveDeployment *Deployment

	LastDeployedAt time.Time
	CreatedAt      time.Time
}

type Resources struct {
	CPU    resource.Quantity
	Memory resource.Quantity
}

type Endpoint struct {
	Host     string
	Port     int
	Protocol Protocol
}

type Protocol string

const (
	ProtocolHTTP  Protocol = "http"
	ProtocolHTTPS Protocol = "https"
	ProtocolTCP   Protocol = "tcp"
)

type ReplicaCount struct {
	Desired int
	Ready   int
}

type AutoscalingSettings struct {
	MinReplicas int
	MaxReplicas int
	TargetCPU   int // percent
}

type ServiceStatus string

const (
	ServiceHealthy   ServiceStatus = "healthy"
	ServiceDegraded  ServiceStatus = "degraded" // some replicas not ready
	ServiceDeploying ServiceStatus = "deploying"
	ServiceFailed    ServiceStatus = "failed"  // no working replicas
	ServiceStopped   ServiceStatus = "stopped" // intentionally scaled to 0
)

type ServiceID struct {
	Workspace   string
	Project     string
	Environment string
	Name        string
}

func ParseServiceID(s string) (ServiceID, error) {
	parts := strings.SplitN(s, "/", 4)

	if len(parts) != 4 || slices.Contains(parts, "") {
		return ServiceID{}, fmt.Errorf("invalid service id %q", s)
	}

	return ServiceID{
		Workspace: parts[0], Project: parts[1],
		Environment: parts[2], Name: parts[3],
	}, nil
}

func (s ServiceID) String() string {
	return s.Workspace + "/" + s.Project + "/" + s.Environment + "/" + s.Name
}

func (s ServiceID) EnvironmentID() EnvironmentID {
	return EnvironmentID{Workspace: s.Workspace, Project: s.Project, Name: s.Environment}
}

func (s ServiceID) Namespace() string {
	return s.EnvironmentID().Namespace()
}

func (s ServiceID) WorkspaceID() string {
	return s.Workspace
}

func (s *ServiceID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return fmt.Errorf("ServiceID must be a string")
	}

	parsed, err := ParseServiceID(str)

	if err != nil {
		return err
	}

	*s = parsed

	return nil
}

func (s ServiceID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(s.String()).MarshalGQL(w)
}

var (
	_ WorkspaceScoped     = (*ServiceID)(nil)
	_ graphql.Marshaler   = ServiceID{}
	_ graphql.Unmarshaler = (*ServiceID)(nil)
)
