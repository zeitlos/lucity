package kubernetes

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/zeitlos/lucity/services/conductor/internal/deployjob"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func (c *Client) Start(ctx context.Context, opts deployjob.StartOptions) (*deployjob.Job, error) {
	if opts.BuildName == "" {
		return nil, errors.New("build name cannot be empty")
	}

	name := "deploy-" + strings.TrimPrefix(opts.BuildName, "build-")

	job := c.newDeployJob(name, opts)
	job, err := c.kubernetes.BatchV1().Jobs(c.config.Namespace).Create(ctx, job, meta.CreateOptions{})

	if apierrors.IsAlreadyExists(err) {
		return c.Get(ctx, deployjob.DeployID{Workspace: opts.Service.Workspace, Name: name})
	}

	if err != nil {
		return nil, err
	}

	return new(toJob(*job)), nil
}

func (c *Client) newDeployJob(name string, opts deployjob.StartOptions) *batch.Job {
	env := []core.EnvVar{
		{Name: "DEPLOY_WORKSPACE", Value: opts.Service.Workspace},
		{Name: "DEPLOY_PROJECT", Value: opts.Service.Project},
		{Name: "DEPLOY_ENVIRONMENT", Value: opts.Service.Environment},
		{Name: "DEPLOY_SERVICE", Value: opts.Service.Name},
		{Name: "DEPLOY_BUILD_NAME", Value: opts.BuildName},
		{Name: "DEPLOY_BUILD_NAMESPACE", Value: c.config.BuildNamespace},
		{Name: "DEPLOY_COMMIT_MESSAGE", Value: opts.CommitMessage},
		{Name: "DEPLOY_RELEASE_ID", Value: opts.ReleaseID},
		{Name: "DEPLOY_RELEASE_TRIGGER", Value: opts.ReleaseTrigger},
		{Name: "DEPLOY_RELEASE_ACTOR", Value: opts.ReleaseActor},
		{Name: "REGISTRY_PULL_URL", Value: c.config.RegistryPullURL},
		{Name: "GATEWAY_NAME", Value: c.config.GatewayName},
		{Name: "GATEWAY_NAMESPACE", Value: c.config.GatewayNS},
		{Name: "CUSTOM_DOMAIN_CLUSTER_ISSUER", Value: c.config.ClusterIssuer},
		{Name: "DATABASE_BACKUP_ENABLED", Value: strconv.FormatBool(c.config.Backups.Enabled)},
		{Name: "DATABASE_BACKUP_S3_ENDPOINT", Value: c.config.Backups.Endpoint},
		{Name: "DATABASE_BACKUP_S3_BUCKET", Value: c.config.Backups.Bucket},
	}

	labelSet := deployJobLabels(opts.Service, opts.ReleaseID)

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
			BackoffLimit:            ptr.To(int32(2)),
			TTLSecondsAfterFinished: ptr.To(int32(7 * 24 * 3600)),
			ActiveDeadlineSeconds:   ptr.To(int64(40 * 60)),
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{Labels: labelSet},
				Spec: core.PodSpec{
					RestartPolicy:      core.RestartPolicyNever,
					ServiceAccountName: c.config.ServiceAccount,
					SecurityContext: &core.PodSecurityContext{
						RunAsNonRoot:   ptr.To(true),
						RunAsUser:      ptr.To(int64(65534)),
						SeccompProfile: &core.SeccompProfile{Type: core.SeccompProfileTypeRuntimeDefault},
					},
					Containers: []core.Container{{
						Name:    "deploy",
						Image:   c.config.Image,
						Command: []string{"/app", "run-deploy"},
						Env:     env,
						Resources: core.ResourceRequirements{
							Requests: core.ResourceList{
								core.ResourceCPU:    resource.MustParse("50m"),
								core.ResourceMemory: resource.MustParse("64Mi"),
							},
							Limits: core.ResourceList{
								core.ResourceCPU:    resource.MustParse("500m"),
								core.ResourceMemory: resource.MustParse("256Mi"),
							},
						},
						SecurityContext: &core.SecurityContext{
							AllowPrivilegeEscalation: ptr.To(false),
							Capabilities:             &core.Capabilities{Drop: []core.Capability{"ALL"}},
						},
					}},
				},
			},
		},
	}
}
