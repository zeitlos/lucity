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
	http  bool
	https bool
}

func (c *Client) addListener(ctx context.Context, hostname, protocol, secretName string) error {
	name := resourceNameFor(hostname)

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

	for _, raw := range listeners {
		entry, ok := raw.(map[string]any)

		if !ok {
			continue
		}

		if entry["name"] == listenerName {
			return nil
		}
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

func (c *Client) removeListener(ctx context.Context, listenerName string) error {
	gw, err := c.dyn.Resource(gatewayGVR).Namespace(c.gatewayNamespace).Get(ctx, c.gatewayName, metav1.GetOptions{})

	if err != nil {
		return err
	}

	listeners, _, _ := unstructured.NestedSlice(gw.Object, "spec", "listeners")
	idx := -1

	for i, raw := range listeners {
		entry, ok := raw.(map[string]any)

		if !ok {
			continue
		}

		if entry["name"] == listenerName {
			idx = i
			break
		}
	}

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

		if protocol == "HTTP" {
			state.http = true
		} else if protocol == "HTTPS" {
			state.https = true
		}

		result[hostname] = state
	}

	return result, nil
}
