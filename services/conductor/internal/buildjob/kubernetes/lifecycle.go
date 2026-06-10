package kubernetes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	hash := c.buildHash(opts.Workspace, *parsed, opts.ContextPath, opts.Commit)
	id := "build-" + hash

	existing, err := c.kubernetes.BatchV1().Jobs(c.namespace).Get(ctx, id, meta.GetOptions{})

	if err == nil {
		b := toJob(*existing)
		return &b, nil
	}

	if !apierrors.IsNotFound(err) {
		return nil, err
	}

	job := c.newBuildJob(id, opts.Workspace, *parsed, opts.ContextPath, opts.Commit, opts.Token, tag, opts.TargetImageNames)

	created, err := c.kubernetes.BatchV1().Jobs(c.namespace).Create(ctx, job, meta.CreateOptions{})

	if err != nil {
		return nil, fmt.Errorf("create build: %w", err)
	}

	return new(toJob(*created)), nil
}

func (c *Client) Cancel(ctx context.Context, id string) (*buildjob.Job, error) {
	job, err := c.kubernetes.BatchV1().Jobs(c.namespace).Get(ctx, id, meta.GetOptions{})

	if err != nil {
		return nil, err
	}

	status := buildStatus(*job)
	terminalStatuses := []buildjob.Status{
		buildjob.StatusCancelled,
		buildjob.StatusCancelling,
		buildjob.StatusFailed,
		buildjob.StatusSucceeded,
	}

	if slices.Contains(terminalStatuses, status) {
		return new(toJob(*job)), nil
	}

	if job.Annotations == nil {
		job.Annotations = map[string]string{}
	}

	job.Annotations[annotationCancelledAt] = time.Now().UTC().Format(time.RFC3339)
	job.Spec.ActiveDeadlineSeconds = new(int64(1))

	job, err = c.kubernetes.BatchV1().Jobs(c.namespace).Update(ctx, job, meta.UpdateOptions{})

	if err != nil {
		return nil, fmt.Errorf("cancel build: %w", err)
	}

	return new(toJob(*job)), nil
}

func (c *Client) newBuildJob(id string, workspaceID string, repoURL url.URL, contextPath, commit, githubToken, tag string, targetImageNames []string) *batch.Job {
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

	labels := map[string]string{
		labelWorkspace:         workspaceID,
		labelRepoHash:          repoURLHash(repoURL),
		labelContextHash:       contextHash(contextPath),
		labelSourceCommit:      commit,
		"lucity.dev/component": "build",
	}

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
						Name:    "build",
						Image:   c.buildRunnerImage,
						Command: []string{"/app", "run-build"},
						Env:     env,
						VolumeMounts: []core.VolumeMount{{
							Name: "registry-auth", MountPath: "/etc/registry-auth", ReadOnly: true,
						}},
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
					Volumes: []core.Volume{
						{
							Name: "work", VolumeSource: core.VolumeSource{EmptyDir: &core.EmptyDirVolumeSource{}},
						},
						{
							Name: "registry-auth",
							VolumeSource: core.VolumeSource{
								Secret: &core.SecretVolumeSource{SecretName: c.registryAuthSecret},
							},
						},
					},
				},
			},
		},
	}
}

func (c *Client) buildHash(workspaceID string, repoURL url.URL, contextPath, commit string) string {
	id := struct {
		W, R, C, Sha, Img string
	}{
		W:   workspaceID,
		R:   normalizeRepoURL(repoURL),
		C:   normalizeContextPath(contextPath),
		Sha: strings.ToLower(strings.TrimSpace(commit)),
		Img: c.buildRunnerImage,
	}

	bytes, _ := json.Marshal(id)
	hash := sha256.Sum256(bytes)

	return hex.EncodeToString(hash[:8])
}
