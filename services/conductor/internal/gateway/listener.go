package gateway

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

type listenerState struct {
	http        bool
	https       bool
	httpsSecret string
}

func findListener(listeners []any, hostname, protocol string) int {
	for i, raw := range listeners {
		entry, ok := raw.(map[string]any)

		if !ok {
			continue
		}

		name, _ := entry["name"].(string)

		if !isManagedListener(name) {
			continue
		}

		if entry["hostname"] == hostname && entry["protocol"] == protocol {
			return i
		}
	}

	return -1
}

func (c *Client) addListener(ctx context.Context, hostname, protocol, secretName string) error {
	name := ResourceNameFor(hostname)

	var listenerName string

	if protocol == "HTTPS" {
		listenerName = name + "-https"
	} else {
		listenerName = name + "-http"
	}

	gw, err := c.dyn.Resource(gatewayGVR).Namespace(c.gatewayNamespace).Get(ctx, c.gatewayName, metav1.GetOptions{})

	if err != nil {
		return err
	}

	listeners, _, _ := unstructured.NestedSlice(gw.Object, "spec", "listeners")

	if findListener(listeners, hostname, protocol) != -1 {
		return nil
	}

	newListener := map[string]any{
		"name":     listenerName,
		"hostname": hostname,
		"protocol": protocol,
		"port":     int64(443),
		"allowedRoutes": map[string]any{
			"namespaces": map[string]any{"from": "All"},
		},
	}

	if protocol == "HTTP" {
		newListener["port"] = int64(80)
	} else {
		newListener["tls"] = map[string]any{
			"mode": "Terminate",
			"certificateRefs": []any{
				map[string]any{"name": secretName},
			},
		}
	}

	patchOps := []map[string]any{
		{"op": "add", "path": "/spec/listeners/-", "value": newListener},
	}

	patchData, err := json.Marshal(patchOps)

	if err != nil {
		return err
	}

	_, err = c.dyn.Resource(gatewayGVR).Namespace(c.gatewayNamespace).Patch(
		ctx, c.gatewayName, types.JSONPatchType, patchData,
		metav1.PatchOptions{FieldManager: "conductor"},
	)

	return err
}

func (c *Client) removeListener(ctx context.Context, hostname, protocol string) error {
	gw, err := c.dyn.Resource(gatewayGVR).Namespace(c.gatewayNamespace).Get(ctx, c.gatewayName, metav1.GetOptions{})

	if err != nil {
		return err
	}

	listeners, _, _ := unstructured.NestedSlice(gw.Object, "spec", "listeners")
	idx := findListener(listeners, hostname, protocol)

	if idx == -1 {
		return nil
	}

	patchOps := []map[string]any{
		{"op": "remove", "path": fmt.Sprintf("/spec/listeners/%d", idx)},
	}

	patchData, err := json.Marshal(patchOps)

	if err != nil {
		return err
	}

	_, err = c.dyn.Resource(gatewayGVR).Namespace(c.gatewayNamespace).Patch(
		ctx, c.gatewayName, types.JSONPatchType, patchData,
		metav1.PatchOptions{FieldManager: "conductor"},
	)

	return err
}

func (c *Client) listListeners(ctx context.Context) (map[string]listenerState, error) {
	gw, err := c.dyn.Resource(gatewayGVR).Namespace(c.gatewayNamespace).Get(ctx, c.gatewayName, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	listeners, _, _ := unstructured.NestedSlice(gw.Object, "spec", "listeners")
	result := make(map[string]listenerState)

	for _, raw := range listeners {
		entry, ok := raw.(map[string]any)

		if !ok {
			continue
		}

		name, _ := entry["name"].(string)

		if !isManagedListener(name) {
			continue
		}

		hostname, _ := entry["hostname"].(string)
		protocol, _ := entry["protocol"].(string)
		state := result[hostname]

		switch protocol {
		case "HTTP":
			state.http = true
		case "HTTPS":
			state.https = true

			refs, _, _ := unstructured.NestedSlice(entry, "tls", "certificateRefs")

			if len(refs) > 0 {
				if ref, ok := refs[0].(map[string]any); ok {
					state.httpsSecret, _ = ref["name"].(string)
				}
			}
		}

		result[hostname] = state
	}

	return result, nil
}
