package kubernetes

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

	existing, err := c.kubernetes.BatchV1().Jobs(c.namespace).List(ctx, meta.ListOptions{
		LabelSelector: labels.Set(buildJobLabels(opts.Workspace, *parsed, opts.ContextPath, opts.Commit)).String(),
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

	job := c.newBuildJob(id, opts.Workspace, *parsed, opts.ContextPath, opts.Commit, opts.Token, tag, opts.TargetImageNames, secretName)
	job, err = c.kubernetes.BatchV1().Jobs(c.namespace).Create(ctx, job, meta.CreateOptions{})

	if err != nil {
		return nil, err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: job.APIVersion,
				Kind:       job.Kind,
				Name:       job.Name,
				UID:        job.UID,
			}},
		},
		StringData: opts.BuildVars,
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

	if job.Labels[labelWorkspace] != id.Workspace {
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

func (c *Client) newBuildJob(id, workspaceID string, repoURL url.URL, contextPath, commit, githubToken, tag string, targetImageNames []string, varsSecret string) *batch.Job {
	targetImages := make([]string, len(targetImageNames))
	targetRefs := make([]string, len(targetImageNames))

	for i, name := range targetImageNames {
		targetImages[i] = name + ":" + tag
		targetRefs[i] = c.registry + "/" + name + ":" + tag
	}

	env := []core.EnvVar{
		{Name: "BUILD_ID", Value: id},
		{Name: "BUILD_SOURCE_URL", Value: repoURL.String()},
		{Name: "BUILD_GIT_REF", Value: commit},
		{Name: "BUILD_CONTEXT_PATH", Value: contextPath},
		{Name: "BUILD_TARGET_REFS", Value: strings.Join(targetRefs, ",")},
		{Name: "BUILDKIT_ADDR", Value: c.buildKitAddr},
		{Name: "GITHUB_TOKEN", Value: githubToken},
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
		volumeMounts = append(volumeMounts, core.VolumeMount{Name: "build-vars", MountPath: "/etc/lucity/build-vars", ReadOnly: true})
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

	labels := buildJobLabels(workspaceID, repoURL, contextPath, commit)

	return &batch.Job{
		ObjectMeta: meta.ObjectMeta{
			Name:      id,
			Namespace: c.namespace,
			Labels:    labels,
			Annotations: map[string]string{
				annotationSourceRepo: repoURL.String(),
				annotationContext:    contextPath,
				annotationTargets:    strings.Join(targetImages, ","),
			},
		},
		Spec: batch.JobSpec{
			BackoffLimit:            ptr.To(int32(0)),
			TTLSecondsAfterFinished: ptr.To(int32(7 * 24 * 3600)),
			ActiveDeadlineSeconds:   ptr.To(int64(30 * 60)),
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{Labels: labels},
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

func buildJobLabels(workspaceID string, repoURL url.URL, contextPath, commit string) map[string]string {
	return map[string]string{
		labelWorkspace:         workspaceID,
		labelRepoHash:          repoURLHash(repoURL),
		labelContextHash:       contextHash(contextPath),
		labelSourceCommit:      commit,
		"lucity.dev/component": "build",
	}
}
