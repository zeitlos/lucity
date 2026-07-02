package kubernetes

import (
	"context"
	"encoding/json"
	"time"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/services/conductor/internal/jobs"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
	"github.com/zeitlos/lucity/services/conductor/internal/scanjob"
)

const (
	labelComponent = labels.Prefix + "component"
	labelScanner   = labels.Prefix + "scanner"
	componentScan  = "scan"

	annotationBuildName = labels.Prefix + "build-name"
)

type Config struct {
	Namespace          string
	Image              string
	Registry           string
	RegistryAuthSecret string
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

var _ scanjob.Interface = (*Client)(nil)

func (c *Client) toJob(ctx context.Context, job batch.Job) scanjob.Job {
	scan := scanjob.Job{
		ID:      scanjob.ScanID{Workspace: job.Labels[labels.Workspace], Name: job.Name},
		Scanner: job.Labels[labelScanner],
		Status:  scanStatus(job),
		Service: platform.ServiceID{
			Workspace:   job.Labels[labels.Workspace],
			Project:     job.Labels[labels.Project],
			Environment: job.Labels[labels.Environment],
			Name:        job.Labels[labels.Service],
		},
		ReleaseID: job.Labels[labels.Release],
		BuildName: job.Annotations[annotationBuildName],
		CreatedAt: job.CreationTimestamp.Time,
	}

	if job.Status.StartTime != nil {
		scan.StartedAt = new(job.Status.StartTime.Time)
	}

	if job.Status.CompletionTime != nil {
		scan.FinishedAt = new(job.Status.CompletionTime.Time)
	} else if scan.Status == scanjob.StatusFailed {
		scan.FinishedAt = failureTime(job)
	}

	if scan.Status == scanjob.StatusSucceeded || scan.Status == scanjob.StatusFailed {
		scan.FindingsCount = c.findingsCount(ctx, job.Name)
	}

	return scan
}

type scanSummary struct {
	Scanner  string `json:"scanner"`
	Findings int    `json:"findings"`
}

func (c *Client) findingsCount(ctx context.Context, jobName string) *int {
	message, err := jobs.TerminationMessage(ctx, c.kubernetes, c.config.Namespace, jobName)

	if err != nil || message == "" {
		return nil
	}

	var summary scanSummary

	if err := json.Unmarshal([]byte(message), &summary); err != nil {
		return nil
	}

	return &summary.Findings
}

func scanStatus(job batch.Job) scanjob.Status {
	for _, condition := range job.Status.Conditions {
		if condition.Status != core.ConditionTrue {
			continue
		}

		switch condition.Type {
		case batch.JobComplete:
			return scanjob.StatusSucceeded
		case batch.JobFailed:
			return scanjob.StatusFailed
		}
	}

	if job.Status.Active > 0 {
		return scanjob.StatusRunning
	}

	return scanjob.StatusQueued
}

func failureTime(job batch.Job) *time.Time {
	for _, condition := range job.Status.Conditions {
		if condition.Status == core.ConditionTrue && condition.Type == batch.JobFailed {
			return new(condition.LastTransitionTime.Time)
		}
	}

	return nil
}

func scanJobLabels(service platform.ServiceID, scanner, releaseID string) map[string]string {
	set := map[string]string{
		labels.Workspace:   service.Workspace,
		labels.Project:     service.Project,
		labels.Environment: service.Environment,
		labels.Service:     service.Name,
		labelComponent:     componentScan,
	}

	if scanner != "" {
		set[labelScanner] = scanner
	}

	if releaseID != "" {
		set[labels.Release] = releaseID
	}

	return set
}
