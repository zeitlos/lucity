package kubernetes

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const probeGracePeriod = time.Minute

const neverReady = "container is running but never became ready"

const (
	readinessProbeFailed = "Readiness probe failed: "
	startupProbeFailed   = "Startup probe failed: "
)

var probeStatusCode = regexp.MustCompile(`statuscode: (\d+)`)

func (c *Client) probeFailures(ctx context.Context, namespace string, pods []core.Pod) map[types.UID]string {
	stalled := make(map[types.UID]bool)

	for _, pod := range pods {
		if stalledPod(pod) {
			stalled[pod.UID] = true
		}
	}

	if len(stalled) == 0 {
		return nil
	}

	events, err := c.kubernetes.CoreV1().Events(namespace).List(ctx, meta.ListOptions{
		FieldSelector: fields.Set{"involvedObject.kind": "Pod", "reason": "Unhealthy"}.String(),
	})

	if err != nil {
		slog.WarnContext(ctx, "failed to list probe events", "error", err, "namespace", namespace)

		return nil
	}

	latest := make(map[types.UID]core.Event)

	for _, event := range events.Items {
		uid := event.InvolvedObject.UID

		if !stalled[uid] || !probeFailed(event.Message) {
			continue
		}

		if current, ok := latest[uid]; ok && !eventTime(event).After(eventTime(current)) {
			continue
		}

		latest[uid] = event
	}

	result := make(map[types.UID]string, len(latest))

	for uid, event := range latest {
		result[uid] = event.Message
	}

	return result
}

func stalledPod(pod core.Pod) bool {
	if pod.DeletionTimestamp != nil || podReady(pod) {
		return false
	}

	for _, container := range pod.Status.ContainerStatuses {
		if running := container.State.Running; running != nil && time.Since(running.StartedAt.Time) >= probeGracePeriod {
			return true
		}
	}

	return false
}

func probeFailed(message string) bool {
	return strings.HasPrefix(message, readinessProbeFailed) || strings.HasPrefix(message, startupProbeFailed)
}

func eventTime(event core.Event) time.Time {
	if !event.LastTimestamp.IsZero() {
		return event.LastTimestamp.Time
	}

	if !event.EventTime.IsZero() {
		return event.EventTime.Time
	}

	return event.CreationTimestamp.Time
}

func probeDiagnosis(pods []core.Pod, failures map[types.UID]string) string {
	for _, pod := range pods {
		if failure, ok := failures[pod.UID]; ok {
			return probeMessage(pod, failure)
		}
	}

	return ""
}

func probeMessage(pod core.Pod, failure string) string {
	containers := pod.Spec.Containers

	if len(containers) == 0 {
		return withDetail(neverReady, failure)
	}

	container := containers[0]
	probe := container.ReadinessProbe
	detail := strings.TrimPrefix(failure, readinessProbeFailed)

	if strings.HasPrefix(failure, startupProbeFailed) {
		probe = container.StartupProbe
		detail = strings.TrimPrefix(failure, startupProbeFailed)
	}

	port := probePort(container, probe)

	if probe == nil || port == 0 {
		return withDetail(neverReady, detail)
	}

	switch {
	case strings.Contains(detail, "connection refused"):
		if portVariable(container) == strconv.Itoa(port) {
			return fmt.Sprintf("nothing is listening on port %d, the port the app gets as PORT", port)
		}

		return fmt.Sprintf("nothing is listening on port %d", port)

	case probeStatusCode.MatchString(detail) && probe.HTTPGet != nil:
		code := probeStatusCode.FindStringSubmatch(detail)[1]

		return fmt.Sprintf("health check %s on port %d answers with HTTP %s", probe.HTTPGet.Path, port, code)

	case strings.Contains(strings.ToLower(detail), "timeout") || strings.Contains(detail, "deadline exceeded"):
		return fmt.Sprintf("port %d does not answer within %ds", port, probe.TimeoutSeconds)
	}

	return withDetail(neverReady, detail)
}

func probePort(container core.Container, probe *core.Probe) int {
	if probe == nil {
		return 0
	}

	var port intstr.IntOrString

	switch {
	case probe.HTTPGet != nil:
		port = probe.HTTPGet.Port
	case probe.TCPSocket != nil:
		port = probe.TCPSocket.Port
	default:
		return 0
	}

	if port.Type == intstr.Int {
		return port.IntValue()
	}

	for _, containerPort := range container.Ports {
		if containerPort.Name == port.StrVal {
			return int(containerPort.ContainerPort)
		}
	}

	return 0
}

func portVariable(container core.Container) string {
	for _, variable := range container.Env {
		if variable.Name == "PORT" && variable.ValueFrom == nil {
			return variable.Value
		}
	}

	return ""
}
