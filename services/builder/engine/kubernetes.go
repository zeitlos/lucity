package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// KubernetesEngine builds images by creating K8s Jobs with a BuildKit sidecar.
// Each build runs in its own Job pod: the main container (build runner) clones
// the repo, generates a railpack plan, and uses buildctl to build+push the image
// via the BuildKit sidecar. Build state is stored in K8s Job status and annotations.
type KubernetesEngine struct {
	client       kubernetes.Interface
	namespace    string
	buildImage   string            // container image for build Jobs (same as builder service)
	nodeSelector map[string]string // optional: schedule builds on specific nodes
	registryURL  string            // internal registry URL for pushing images
	insecure     bool              // allow HTTP registry
}

// KubernetesEngineOpts configures the KubernetesEngine.
type KubernetesEngineOpts struct {
	Client       kubernetes.Interface
	Namespace    string
	BuildImage   string
	NodeSelector map[string]string
	RegistryURL  string
	Insecure     bool
}

// NewKubernetesEngine creates a KubernetesEngine.
func NewKubernetesEngine(opts KubernetesEngineOpts) *KubernetesEngine {
	return &KubernetesEngine{
		client:       opts.Client,
		namespace:    opts.Namespace,
		buildImage:   opts.BuildImage,
		nodeSelector: opts.NodeSelector,
		registryURL:  opts.RegistryURL,
		insecure:     opts.Insecure,
	}
}

func (e *KubernetesEngine) Detect(ctx context.Context, repoPath string) ([]DetectResult, error) {
	return Detect(ctx, repoPath)
}

func (e *KubernetesEngine) Build(ctx context.Context, opts BuildOpts) (*BuildResult, error) {
	jobName := "build-" + opts.BuildID[:8]

	job := e.buildJob(jobName, opts)

	slog.Info("creating build job", "job", jobName, "build_id", opts.BuildID, "image", opts.ImageName)

	created, err := e.client.BatchV1().Jobs(e.namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create build job: %w", err)
	}

	// Poll until the Job completes
	result, err := e.waitForJob(ctx, created.Name, opts.BuildID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// buildJob constructs the K8s Job spec for a build.
func (e *KubernetesEngine) buildJob(name string, opts BuildOpts) *batchv1.Job {
	backoffLimit := int32(0)
	ttl := int32(86400) // 24 hours

	// BuildKit socket address
	buildkitAddr := "unix:///run/buildkit/buildkitd.sock"

	// Environment variables for the build runner
	env := []corev1.EnvVar{
		{Name: "BUILD_ID", Value: opts.BuildID},
		{Name: "BUILD_SOURCE_URL", Value: opts.SourceURL},
		{Name: "BUILD_GIT_REF", Value: opts.GitRef},
		{Name: "BUILD_REGISTRY", Value: opts.Registry},
		{Name: "BUILD_CONTEXT_PATH", Value: opts.ContextPath},
		{Name: "BUILD_INSECURE", Value: fmt.Sprintf("%t", opts.Insecure)},
		{Name: "BUILDKIT_ADDR", Value: buildkitAddr},
		{Name: "GITHUB_TOKEN", Value: opts.GitHubToken},
		{Name: "BUILD_NAMESPACE", Value: e.namespace},
	}

	// Security context for buildkitd rootless
	privileged := false
	seccompUnconfined := corev1.SeccompProfile{
		Type: corev1.SeccompProfileTypeUnconfined,
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: e.namespace,
			Labels: map[string]string{
				"lucity.dev/build-id":  opts.BuildID,
				"lucity.dev/component": "build",
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttl,
			BackoffLimit:            &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"lucity.dev/build-id":  opts.BuildID,
						"lucity.dev/component": "build",
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:       e.nodeSelector,
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: "lucity-builder",
					Containers: []corev1.Container{
						{
							Name:    "build",
							Image:   e.buildImage,
							Command: []string{"/app", "run-build"},
							Env:     env,
							VolumeMounts: []corev1.VolumeMount{
								{Name: "buildkit-socket", MountPath: "/run/buildkit"},
								{Name: "work", MountPath: "/tmp/lucity-builds"},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    mustParseQuantity("500m"),
									corev1.ResourceMemory: mustParseQuantity("512Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    mustParseQuantity("2"),
									corev1.ResourceMemory: mustParseQuantity("2Gi"),
								},
							},
						},
						{
							Name:  "buildkitd",
							Image: "moby/buildkit:rootless",
							Args: []string{
								"--addr", "unix:///run/buildkit/buildkitd.sock",
								"--oci-worker-no-process-sandbox",
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged:     &privileged,
								SeccompProfile: &seccompUnconfined,
							},
							Env: []corev1.EnvVar{
								{Name: "BUILDKIT_HOST", Value: "unix:///run/buildkit/buildkitd.sock"},
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "buildkit-socket", MountPath: "/run/buildkit"},
								{Name: "buildkit-config", MountPath: "/etc/buildkit", ReadOnly: true},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"buildctl", "debug", "workers"},
									},
								},
								InitialDelaySeconds: 2,
								PeriodSeconds:       2,
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "buildkit-socket",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "work",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "buildkit-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "lucity-buildkit",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// waitForJob polls the Job until it reaches a terminal state and returns the build result.
func (e *KubernetesEngine) waitForJob(ctx context.Context, jobName, buildID string) (*BuildResult, error) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			job, err := e.client.BatchV1().Jobs(e.namespace).Get(ctx, jobName, metav1.GetOptions{})
			if err != nil {
				slog.Warn("failed to get build job", "job", jobName, "error", err)
				continue
			}

			for _, c := range job.Status.Conditions {
				if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
					return e.readResult(job)
				}
				if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
					errMsg := "build job failed"
					if ann, ok := job.Annotations["lucity.dev/error"]; ok {
						errMsg = ann
					}
					return nil, fmt.Errorf("%s", errMsg)
				}
			}
		}
	}
}

// readResult reads the build result from Job annotations.
func (e *KubernetesEngine) readResult(job *batchv1.Job) (*BuildResult, error) {
	ann, ok := job.Annotations["lucity.dev/result"]
	if !ok {
		return nil, fmt.Errorf("build job completed but no result annotation found")
	}

	var result struct {
		ImageRef string `json:"imageRef"`
		Digest   string `json:"digest"`
	}
	if err := json.Unmarshal([]byte(ann), &result); err != nil {
		return nil, fmt.Errorf("failed to parse build result: %w", err)
	}

	return &BuildResult{
		ImageRef: result.ImageRef,
		Digest:   result.Digest,
	}, nil
}

func mustParseQuantity(s string) resource.Quantity {
	q, err := resource.ParseQuantity(s)
	if err != nil {
		panic(fmt.Sprintf("invalid quantity %q: %v", s, err))
	}
	return q
}
