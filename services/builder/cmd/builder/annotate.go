package main

// Build-result annotations on K8s Jobs are the contract between this
// in-pod build runner and the conductor's build orchestrator
// (services/conductor/internal/builds). The conductor reads
// `lucity.dev/result` (success: image ref + digest) and
// `lucity.dev/error` (failure: free-text message) to determine the
// outcome of a Build Job.
//
// Keep the annotation keys and the JSON shape stable across both
// sides — they are the only mechanism by which build state crosses
// from the runner pod back to the platform.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	annotationResult = "lucity.dev/result"
	annotationError  = "lucity.dev/error"
)

type buildResult struct {
	ImageRef string `json:"imageRef"`
	Digest   string `json:"digest"`
}

// annotateJobResult records a successful build's image ref + digest
// on the parent Job. The conductor's tracker reads this annotation
// to surface the build outcome.
func annotateJobResult(client kubernetes.Interface, namespace, buildID, imageRef, digest string) error {
	job, err := findJobByBuildID(client, namespace, buildID)
	if err != nil {
		return err
	}

	res, err := json.Marshal(buildResult{ImageRef: imageRef, Digest: digest})
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	// K8s annotation values are strings; double-encode the JSON.
	resStr, err := json.Marshal(string(res))
	if err != nil {
		return fmt.Errorf("encode result string: %w", err)
	}

	patch := fmt.Sprintf(`{"metadata":{"annotations":{%q:%s}}}`, annotationResult, string(resStr))
	_, err = client.BatchV1().Jobs(namespace).Patch(
		context.Background(),
		job.Name,
		"application/merge-patch+json",
		[]byte(patch),
		metav1.PatchOptions{},
	)
	return err
}

// annotateJobError records a failure message on the parent Job.
func annotateJobError(client kubernetes.Interface, namespace, buildID, errMsg string) error {
	job, err := findJobByBuildID(client, namespace, buildID)
	if err != nil {
		return err
	}

	escaped := strings.ReplaceAll(errMsg, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	patch := fmt.Sprintf(`{"metadata":{"annotations":{%q:"%s"}}}`, annotationError, escaped)
	_, err = client.BatchV1().Jobs(namespace).Patch(
		context.Background(),
		job.Name,
		"application/merge-patch+json",
		[]byte(patch),
		metav1.PatchOptions{},
	)
	return err
}

// findJobByBuildID locates the parent Job via its lucity.dev/build-id
// label. The conductor sets this label when it creates the Job.
func findJobByBuildID(client kubernetes.Interface, namespace, buildID string) (*batchv1.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobs, err := client.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("lucity.dev/build-id=%s", buildID),
	})
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	if len(jobs.Items) == 0 {
		return nil, fmt.Errorf("job not found for build %s", buildID)
	}
	return &jobs.Items[0], nil
}
