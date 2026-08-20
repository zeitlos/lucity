package kubernetes

import (
	"time"

	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/services/conductor/internal/deployjob"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	labelComponent  = labels.Prefix + "component"
	componentDeploy = "deploy"

	annotationBuildName = labels.Prefix + "build-name"
)

type Config struct {
	Namespace       string
	Image           string
	ServiceAccount  string
	BuildNamespace  string
	RegistryPullURL string
	GatewayName     string
	GatewayNS       string
	ClusterIssuer   string
	Backups         BackupConfig
}

type BackupConfig struct {
	Enabled  bool
	Endpoint string
	Bucket   string
}

type Client struct {
	config     Config
	kubernetes kubernetes.Interface
}

func New(kubernetes kubernetes.Interface, config Config) *Client {
	return &Client{
		config:     config,
		kubernetes: kubernetes,
	}
}

var _ deployjob.Interface = (*Client)(nil)

func toJob(job batch.Job) deployjob.Job {
	deploy := deployjob.Job{
		ID:     deployjob.DeployID{Workspace: job.Labels[labels.Workspace], Name: job.Name},
		Status: deployStatus(job),
		Service: platform.ServiceID{
			Workspace:   job.Labels[labels.Workspace],
			Project:     job.Labels[labels.Project],
			Environment: job.Labels[labels.Environment],
			Name:        job.Labels[labels.Service],
		},
		ReleaseID: job.Labels[labels.Release],
		BuildName: job.Annotations[annotationBuildName],
	}

	if job.Status.StartTime != nil {
		deploy.StartedAt = new(job.Status.StartTime.Time)
	}

	if job.Status.CompletionTime != nil {
		deploy.FinishedAt = new(job.Status.CompletionTime.Time)
	} else if deploy.Status == deployjob.StatusFailed {
		deploy.FinishedAt = failureTime(job)
	}

	return deploy
}

func deployStatus(job batch.Job) deployjob.Status {
	for _, c := range job.Status.Conditions {
		if c.Status != core.ConditionTrue {
			continue
		}

		switch c.Type {
		case batch.JobComplete:
			return deployjob.StatusSucceeded
		case batch.JobFailed:
			return deployjob.StatusFailed
		}
	}

	if job.Status.Active > 0 {
		return deployjob.StatusRunning
	}

	return deployjob.StatusQueued
}

func failureTime(job batch.Job) *time.Time {
	for _, c := range job.Status.Conditions {
		if c.Status == core.ConditionTrue && c.Type == batch.JobFailed {
			return new(c.LastTransitionTime.Time)
		}
	}

	return nil
}

func deployJobLabels(service platform.ServiceID, releaseID string) map[string]string {
	set := map[string]string{
		labels.Workspace:   service.Workspace,
		labels.Project:     service.Project,
		labels.Environment: service.Environment,
		labels.Service:     service.Name,
		labelComponent:     componentDeploy,
	}

	if releaseID != "" {
		set[labels.Release] = releaseID
	}

	return set
}
