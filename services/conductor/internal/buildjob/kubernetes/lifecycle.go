package kubernetes

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/joho/godotenv"
	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/ptr"
)

func (c *Client) Start(ctx context.Context, opts buildjob.StartOptions) (*buildjob.Job, error) {
	if len(opts.TargetImageNames) == 0 {
		return nil, errors.New("no target image names")
	}

	if opts.Commit == "" {
		return nil, errors.New("commit cannot be empty")
	}

	parsed, err := url.Parse(opts.RepoURL)

	if err != nil {
		return nil, err
	}

	tag := opts.Commit

	if len(tag) > 7 {
		tag = tag[:7]
	}

	dedupe := buildJobLabels(opts.Service)
	dedupe[labelSourceCommit] = opts.Commit

	existing, err := c.kubernetes.BatchV1().Jobs(c.namespace).List(ctx, meta.ListOptions{
		LabelSelector: k8slabels.Set(dedupe).String(),
	})

	if err != nil {
		return nil, err
	}

	for _, job := range existing.Items {
		if !isDone(job) {
			return new(toJob(job)), nil
		}
	}

	id := "build-" + rand.String(6)
	secretName := id + "-variables"

	job := c.newBuildJob(id, opts, *parsed, tag, secretName)
	job, err = c.kubernetes.BatchV1().Jobs(c.namespace).Create(ctx, job, meta.CreateOptions{})

	if err != nil {
		return nil, err
	}

	dotEnv, err := godotenv.Marshal(opts.BuildVars)

	if err != nil {
		return nil, err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "batch/v1",
				Kind:       "Job",
				Name:       job.Name,
				UID:        job.UID,
			}},
		},
		StringData: map[string]string{
			".env": dotEnv,
		},
	}

	secret, err = c.kubernetes.CoreV1().Secrets(c.namespace).Create(ctx, secret, meta.CreateOptions{})

	if err != nil {
		return nil, err
	}

	return new(toJob(*job)), nil
}

func (c *Client) Cancel(ctx context.Context, id buildjob.BuildID) (*buildjob.Job, error) {
	job, err := c.kubernetes.BatchV1().Jobs(c.namespace).Get(ctx, id.Name, meta.GetOptions{})

	if err != nil {
		return nil, err
	}

	if job.Labels[labels.Workspace] != id.Workspace {
		return nil, errors.New("build not found")
	}

	if isDone(*job) {
		return new(toJob(*job)), nil
	}

	if job.Annotations == nil {
		job.Annotations = map[string]string{}
	}

	job.Annotations[annotationCancelledAt] = time.Now().UTC().Format(time.RFC3339)
	job.Spec.ActiveDeadlineSeconds = new(int64(1))

	job, err = c.kubernetes.BatchV1().Jobs(c.namespace).Update(ctx, job, meta.UpdateOptions{})

	if err != nil {
		return nil, err
	}

	return new(toJob(*job)), nil
}

func isDone(job batch.Job) bool {
	status := buildStatus(job)
	terminalStatuses := []buildjob.Status{
		buildjob.StatusCancelled,
		buildjob.StatusCancelling,
		buildjob.StatusFailed,
		buildjob.StatusSucceeded,
	}

	return slices.Contains(terminalStatuses, status)
}

func (c *Client) newBuildJob(id string, opts buildjob.StartOptions, repoURL url.URL, tag, varsSecret string) *batch.Job {
	targetImages := make([]string, len(opts.TargetImageNames))
	targetRefs := make([]string, len(opts.TargetImageNames))

	for i, name := range opts.TargetImageNames {
		targetImages[i] = name + ":" + tag
		targetRefs[i] = c.registry + "/" + name + ":" + tag
	}

	env := []core.EnvVar{
		{Name: "BUILD_ID", Value: id},
		{Name: "BUILD_SOURCE_URL", Value: repoURL.String()},
		{Name: "BUILD_GIT_REF", Value: opts.Commit},
		{Name: "BUILD_CONTEXT_PATH", Value: opts.ContextPath},
		{Name: "BUILD_TARGET_REFS", Value: strings.Join(targetRefs, ",")},
		{Name: "BUILDKIT_ADDR", Value: c.buildKitAddr},
		{Name: "GITHUB_TOKEN", Value: opts.Token},
		{Name: "DOCKER_CONFIG", Value: "/etc/registry-auth"},
	}

	volumeMounts := []core.VolumeMount{
		{Name: "registry-auth", MountPath: "/etc/registry-auth", ReadOnly: true},
	}

	volumes := []core.Volume{
		{Name: "work", VolumeSource: core.VolumeSource{EmptyDir: &core.EmptyDirVolumeSource{}}},
		{Name: "registry-auth", VolumeSource: core.VolumeSource{Secret: &core.SecretVolumeSource{SecretName: c.registryAuthSecret}}},
	}

	if varsSecret != "" {
		volumeMounts = append(volumeMounts, core.VolumeMount{Name: "build-vars", MountPath: "/etc/lucity", ReadOnly: true})
		volumes = append(volumes, core.Volume{Name: "build-vars", VolumeSource: core.VolumeSource{Secret: &core.SecretVolumeSource{SecretName: varsSecret}}})
	}

	if c.buildKitTLSSecret != "" {
		env = append(env,
			core.EnvVar{Name: "BUILDKIT_TLS_CA", Value: "/etc/buildkit-certs/ca.crt"},
			core.EnvVar{Name: "BUILDKIT_TLS_CERT", Value: "/etc/buildkit-certs/tls.crt"},
			core.EnvVar{Name: "BUILDKIT_TLS_KEY", Value: "/etc/buildkit-certs/tls.key"},
			core.EnvVar{Name: "BUILDKIT_SERVER_NAME", Value: c.buildKitServerName},
		)

		volumeMounts = append(volumeMounts, core.VolumeMount{
			Name: "buildkit-certs", MountPath: "/etc/buildkit-certs", ReadOnly: true,
		})

		volumes = append(volumes, core.Volume{
			Name:         "buildkit-certs",
			VolumeSource: core.VolumeSource{Secret: &core.SecretVolumeSource{SecretName: c.buildKitTLSSecret}},
		})
	}

	labelSet := buildJobLabels(opts.Service)
	labelSet[labelSourceCommit] = opts.Commit

	if opts.ReleaseID != "" {
		labelSet[labels.Release] = opts.ReleaseID
	}

	annotations := map[string]string{
		annotationSourceRepo: repoURL.String(),
		annotationContext:    opts.ContextPath,
		annotationTargets:    strings.Join(targetImages, ","),
	}

	if opts.CommitMessage != "" {
		annotations[annotationCommitMessage] = truncate(opts.CommitMessage, 1024)
	}

	return &batch.Job{
		ObjectMeta: meta.ObjectMeta{
			Name:        id,
			Namespace:   c.namespace,
			Labels:      labelSet,
			Annotations: annotations,
		},
		Spec: batch.JobSpec{
			Suspend:                 ptr.To(true),
			BackoffLimit:            ptr.To(int32(0)),
			TTLSecondsAfterFinished: ptr.To(int32(7 * 24 * 3600)),
			ActiveDeadlineSeconds:   ptr.To(int64(30 * 60)),
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
						Name:         "build",
						Image:        c.buildRunnerImage,
						Command:      []string{"/app", "run-build"},
						Env:          env,
						VolumeMounts: volumeMounts,
						Resources: core.ResourceRequirements{
							Requests: core.ResourceList{
								core.ResourceCPU:    resource.MustParse("250m"),
								core.ResourceMemory: resource.MustParse("512Mi"),
							},
							Limits: core.ResourceList{
								core.ResourceCPU:    resource.MustParse("2"),
								core.ResourceMemory: resource.MustParse("4Gi"),
							},
						},
						SecurityContext: &core.SecurityContext{
							AllowPrivilegeEscalation: ptr.To(false),
							Capabilities:             &core.Capabilities{Drop: []core.Capability{"ALL"}},
						},
					}},
					Volumes: volumes,
				},
			},
		},
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}

	cut := s[:max]

	for len(cut) > 0 && !utf8.ValidString(cut) {
		cut = cut[:len(cut)-1]
	}

	return cut
}
