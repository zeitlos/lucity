package tests

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// kubectl runs a kubectl command and returns stdout. Fatals on error.
func kubectl(t *testing.T, args ...string) string {
	t.Helper()
	out, err := kubectlRaw(args...)
	if err != nil {
		t.Fatalf("kubectl %s failed: %v\noutput: %s", strings.Join(args, " "), err, out)
	}
	return out
}

// kubectlRaw runs a kubectl command without requiring *testing.T.
func kubectlRaw(args ...string) (string, error) {
	cmd := exec.Command("kubectl", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// kubectlQuiet runs kubectl and returns stdout + error (does not fatal).
func kubectlQuiet(t *testing.T, args ...string) (string, error) {
	t.Helper()
	return kubectlRaw(args...)
}

// kubectlJSON runs kubectl with -o json and decodes into target.
func kubectlJSON(t *testing.T, target any, args ...string) {
	t.Helper()
	fullArgs := append(args, "-o", "json")
	out := kubectl(t, fullArgs...)
	if err := json.Unmarshal([]byte(out), target); err != nil {
		t.Fatalf("failed to decode kubectl JSON output: %v\nraw: %s", err, out)
	}
}

// waitForNamespace polls until the namespace exists or times out. Fatals on timeout.
func waitForNamespace(t *testing.T, ns string, timeout time.Duration) {
	t.Helper()
	if !waitForNamespaceOK(t, ns, timeout) {
		t.Fatalf("namespace %q did not appear within %s", ns, timeout)
	}
}

// waitForNamespaceOK polls until the namespace exists. Returns false on timeout (non-fatal).
func waitForNamespaceOK(t *testing.T, ns string, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, err := kubectlQuiet(t, "get", "namespace", ns, "--no-headers")
		if err == nil {
			t.Logf("namespace %s exists", ns)
			return true
		}
		time.Sleep(2 * time.Second)
	}
	t.Logf("namespace %q did not appear within %s", ns, timeout)
	return false
}

// requireNamespace skips the test if the development namespace isn't ready.
func requireNamespace(t *testing.T) {
	t.Helper()
	if !devNamespaceReady {
		t.Skip("skipping: development namespace not ready (ArgoCD may not have synced)")
	}
}

// waitForNamespaceGone polls until the namespace no longer exists. Fatals on timeout.
func waitForNamespaceGone(t *testing.T, ns string, timeout time.Duration) {
	t.Helper()
	if !waitForNamespaceGoneOK(t, ns, timeout) {
		t.Fatalf("namespace %q still exists after %s", ns, timeout)
	}
}

// waitForNamespaceGoneOK polls until the namespace no longer exists. Returns false on timeout (non-fatal).
func waitForNamespaceGoneOK(t *testing.T, ns string, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		out, err := kubectlQuiet(t, "get", "namespace", ns, "--no-headers")
		if err != nil {
			t.Logf("namespace %s is gone", ns)
			return true
		}
		// Also accept Terminating
		if strings.Contains(out, "Terminating") {
			t.Logf("namespace %s is terminating", ns)
			return true
		}
		time.Sleep(2 * time.Second)
	}
	t.Logf("namespace %q still exists after %s", ns, timeout)
	return false
}

// waitForPod polls until at least one pod matching the label selector is Running.
func waitForPod(t *testing.T, ns, labelSelector string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		out, err := kubectlQuiet(t,
			"get", "pods", "-n", ns,
			"-l", labelSelector,
			"--no-headers",
			"-o", "custom-columns=:status.phase",
		)
		if err == nil && strings.Contains(out, "Running") {
			t.Logf("pod with selector %s is Running in %s", labelSelector, ns)
			return
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("no Running pod with selector %q in namespace %q within %s", labelSelector, ns, timeout)
}

// assertResourceExists verifies a Kubernetes resource exists.
func assertResourceExists(t *testing.T, resource, name, ns string) {
	t.Helper()
	args := []string{"get", resource, name}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	kubectl(t, args...)
	t.Logf("%s/%s exists in %s", resource, name, ns)
}

// assertResourceGone verifies a Kubernetes resource does not exist.
func assertResourceGone(t *testing.T, resource, name, ns string) {
	t.Helper()
	args := []string{"get", resource, name}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	_, err := kubectlQuiet(t, args...)
	if err == nil {
		t.Fatalf("expected %s/%s to be gone in %s, but it still exists", resource, name, ns)
	}
	t.Logf("%s/%s is gone in %s", resource, name, ns)
}

// getDeploymentImage returns the container image of a deployment.
func getDeploymentImage(t *testing.T, ns, deployName string) string {
	t.Helper()
	out := kubectl(t,
		"get", "deployment", deployName, "-n", ns,
		"-o", "jsonpath={.spec.template.spec.containers[0].image}",
	)
	return strings.TrimSpace(out)
}
