package build

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/pkg/builder"
)

// buildResult is the JSON structure stored in the Job annotation.
type buildResult struct {
	ImageRef string `json:"imageRef"`
	Digest   string `json:"digest"`
}

// K8sTracker reads build state from Kubernetes Jobs and pod logs.
// It tracks known build IDs in memory so it can return a QUEUED state
// before the K8s Job is created.
type K8sTracker struct {
	client    kubernetes.Interface
	namespace string
	known     map[string]bool
	mu        sync.RWMutex
}

// NewK8sTracker creates a tracker that reads build state from Kubernetes.
func NewK8sTracker(client kubernetes.Interface, namespace string) *K8sTracker {
	return &K8sTracker{
		client:    client,
		namespace: namespace,
		known:     make(map[string]bool),
	}
}

// Create registers a build ID so we can return QUEUED before the Job exists.
func (t *K8sTracker) Create(id string) {
	t.mu.Lock()
	t.known[id] = true
	t.mu.Unlock()
}

// Get returns the current build state by querying the K8s Job.
// If the build ID is known but the Job doesn't exist yet, returns QUEUED.
func (t *K8sTracker) Get(id string) *BuildState {
	job, err := t.findJob(id)
	if err != nil || job == nil {
		t.mu.RLock()
		isKnown := t.known[id]
		t.mu.RUnlock()

		if isKnown {
			return &BuildState{
				ID:    id,
				Phase: builder.BuildPhase_BUILD_PHASE_QUEUED,
			}
		}
		return nil
	}

	state := &BuildState{
		ID:    id,
		Phase: jobPhase(job),
	}

	// Read result from Job annotation (set by build runner on success)
	if ann, ok := job.Annotations["lucity.dev/result"]; ok {
		var res buildResult
		if err := json.Unmarshal([]byte(ann), &res); err == nil {
			state.ImageRef = res.ImageRef
			state.Digest = res.Digest
		}
	}

	// Read error from annotation (set by build runner on failure)
	if errMsg, ok := job.Annotations["lucity.dev/error"]; ok {
		state.Error = errMsg
	}

	return state
}

// Update is a no-op — phase is derived from Job status.
func (t *K8sTracker) Update(id string, phase builder.BuildPhase) {}

// Succeed is a no-op — the build runner annotates the Job directly.
func (t *K8sTracker) Succeed(id, imageRef, digest string) {}

// Fail stores an error for builds that fail before the Job is created
// (e.g. during clone). For in-Job failures, the build runner annotates directly.
func (t *K8sTracker) Fail(id, errMsg string) {
	t.mu.Lock()
	delete(t.known, id)
	t.mu.Unlock()
}

// AppendLog is a no-op — logs come from pod stdout via K8s API.
func (t *K8sTracker) AppendLog(id, line string) {}

// LogLines returns build log lines from the K8s pod logs API.
func (t *K8sTracker) LogLines(id string, offset int) []string {
	pod, err := t.findBuildPod(id)
	if err != nil || pod == nil {
		return nil
	}

	// Read logs from the "build" container
	logOpts := &corev1.PodLogOptions{
		Container: "build",
	}
	req := t.client.CoreV1().Pods(t.namespace).GetLogs(pod.Name, logOpts)
	stream, err := req.Stream(context.Background())
	if err != nil {
		slog.Debug("failed to stream pod logs", "pod", pod.Name, "error", err)
		return nil
	}
	defer stream.Close()

	var allLines []string
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	if offset >= len(allLines) {
		return nil
	}
	return allLines[offset:]
}

// LogCount returns the total number of log lines.
func (t *K8sTracker) LogCount(id string) int {
	lines := t.LogLines(id, 0)
	return len(lines)
}

// IsTerminal returns true if the Job has completed (succeeded or failed).
func (t *K8sTracker) IsTerminal(id string) bool {
	job, err := t.findJob(id)
	if err != nil || job == nil {
		// If we know about it but no Job exists, it's still pending
		t.mu.RLock()
		isKnown := t.known[id]
		t.mu.RUnlock()
		return !isKnown
	}

	phase := jobPhase(job)
	return phase == builder.BuildPhase_BUILD_PHASE_SUCCEEDED || phase == builder.BuildPhase_BUILD_PHASE_FAILED
}

// findJob finds a Job by build ID label.
func (t *K8sTracker) findJob(buildID string) (*batchv1.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobs, err := t.client.BatchV1().Jobs(t.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("lucity.dev/build-id=%s", buildID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	if len(jobs.Items) == 0 {
		return nil, nil
	}
	return &jobs.Items[0], nil
}

// findBuildPod finds the pod for a build Job.
func (t *K8sTracker) findBuildPod(buildID string) (*corev1.Pod, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pods, err := t.client.CoreV1().Pods(t.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("lucity.dev/build-id=%s", buildID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}
	if len(pods.Items) == 0 {
		return nil, nil
	}
	return &pods.Items[0], nil
}

// jobPhase maps K8s Job status to BuildPhase.
func jobPhase(job *batchv1.Job) builder.BuildPhase {
	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
			return builder.BuildPhase_BUILD_PHASE_SUCCEEDED
		}
		if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
			return builder.BuildPhase_BUILD_PHASE_FAILED
		}
	}

	if job.Status.Active > 0 {
		return builder.BuildPhase_BUILD_PHASE_BUILDING
	}

	return builder.BuildPhase_BUILD_PHASE_QUEUED
}

// AnnotateJobResult annotates a Job with the build result. Called by the build
// runner inside the Job pod to communicate results back to the builder service.
func AnnotateJobResult(client kubernetes.Interface, namespace, buildID, imageRef, digest string) error {
	job, err := findJobByBuildID(client, namespace, buildID)
	if err != nil {
		return err
	}

	res, err := json.Marshal(buildResult{ImageRef: imageRef, Digest: digest})
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	// Annotation values must be strings — encode the JSON as a string value
	resStr, err := json.Marshal(string(res))
	if err != nil {
		return fmt.Errorf("failed to encode result string: %w", err)
	}

	patch := fmt.Sprintf(`{"metadata":{"annotations":{"lucity.dev/result":%s}}}`, string(resStr))
	_, err = client.BatchV1().Jobs(namespace).Patch(
		context.Background(),
		job.Name,
		"application/merge-patch+json",
		[]byte(patch),
		metav1.PatchOptions{},
	)
	return err
}

// AnnotateJobError annotates a Job with an error message.
func AnnotateJobError(client kubernetes.Interface, namespace, buildID, errMsg string) error {
	job, err := findJobByBuildID(client, namespace, buildID)
	if err != nil {
		return err
	}

	// Escape for JSON
	escaped := strings.ReplaceAll(errMsg, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")

	patch := fmt.Sprintf(`{"metadata":{"annotations":{"lucity.dev/error":"%s"}}}`, escaped)
	_, err = client.BatchV1().Jobs(namespace).Patch(
		context.Background(),
		job.Name,
		"application/merge-patch+json",
		[]byte(patch),
		metav1.PatchOptions{},
	)
	return err
}

func findJobByBuildID(client kubernetes.Interface, namespace, buildID string) (*batchv1.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobs, err := client.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("lucity.dev/build-id=%s", buildID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	if len(jobs.Items) == 0 {
		return nil, fmt.Errorf("job not found for build %s", buildID)
	}
	return &jobs.Items[0], nil
}
