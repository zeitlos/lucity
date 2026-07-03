package kubernetes

import (
	"context"
	"errors"
	"strings"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/zeitlos/lucity/services/conductor/internal/scanjob"
)

func (c *Client) Start(ctx context.Context, opts scanjob.StartOptions) (*scanjob.Job, error) {
	if opts.BuildName == "" {
		return nil, errors.New("build name cannot be empty")
	}

	name := "scan-" + strings.TrimPrefix(opts.BuildName, "build-")

	job := c.newScanJob(name, opts)
	job, err := c.kubernetes.BatchV1().Jobs(c.config.Namespace).Create(ctx, job, meta.CreateOptions{})

	if apierrors.IsAlreadyExists(err) {
		return c.Get(ctx, scanjob.ScanID{Workspace: opts.Service.Workspace, Name: name})
	}

	if err != nil {
		return nil, err
	}

	return new(c.toJob(ctx, *job)), nil
}

func (c *Client) newScanJob(name string, opts scanjob.StartOptions) *batch.Job {
	reportRepo := c.config.Registry + "/" + opts.Service.Workspace + "/" + opts.Service.Project + "/" + opts.Service.Name + "/scans"

	env := []core.EnvVar{
		{Name: "SCAN_ID", Value: name},
		{Name: "SCAN_SOURCE_URL", Value: opts.SourceURL},
		{Name: "SCAN_COMMIT", Value: opts.Commit},
		{Name: "SCAN_REPORT_REPO", Value: reportRepo},
		{Name: "GITHUB_TOKEN", Value: opts.Token},
		{Name: "DOCKER_CONFIG", Value: "/etc/registry-auth"},
	}

	labelSet := scanJobLabels(opts.Service, opts.ReleaseID)

	return &batch.Job{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: c.config.Namespace,
			Labels:    labelSet,
			Annotations: map[string]string{
				annotationBuildName: opts.BuildName,
			},
		},
		Spec: batch.JobSpec{
			Suspend:                 ptr.To(true),
			BackoffLimit:            ptr.To(int32(0)),
			TTLSecondsAfterFinished: ptr.To(int32(7 * 24 * 3600)),
			ActiveDeadlineSeconds:   ptr.To(int64(65 * 60)), // 1 hour 5 minutes
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{Labels: labelSet},
				Spec: core.PodSpec{
					HostUsers:                    ptr.To(false),
					RestartPolicy:                core.RestartPolicyNever,
					AutomountServiceAccountToken: ptr.To(false),
					SecurityContext: &core.PodSecurityContext{
						SeccompProfile: &core.SeccompProfile{Type: core.SeccompProfileTypeRuntimeDefault},
					},
					Containers: []core.Container{{
						Name:    "scan",
						Image:   c.config.Image,
						Command: []string{"/app", "run-scan"},
						Env:     env,
						VolumeMounts: []core.VolumeMount{
							{Name: "registry-auth", MountPath: "/etc/registry-auth", ReadOnly: true},
						},
						Resources: core.ResourceRequirements{
							Requests: core.ResourceList{
								core.ResourceCPU:    resource.MustParse("500m"),
								core.ResourceMemory: resource.MustParse("2Gi"),
							},
							Limits: core.ResourceList{
								core.ResourceCPU:    resource.MustParse("4"),
								core.ResourceMemory: resource.MustParse("8Gi"),
							},
						},
						SecurityContext: &core.SecurityContext{
							AllowPrivilegeEscalation: ptr.To(false),
							Capabilities:             &core.Capabilities{Drop: []core.Capability{"ALL"}},
						},
					}},
					Volumes: []core.Volume{
						{Name: "registry-auth", VolumeSource: core.VolumeSource{Secret: &core.SecretVolumeSource{SecretName: c.config.RegistryAuthSecret}}},
					},
				},
			},
		},
	}
}
