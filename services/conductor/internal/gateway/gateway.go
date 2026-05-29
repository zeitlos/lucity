package gateway

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Client struct {
	dyn              dynamic.Interface
	gatewayName      string
	gatewayNamespace string
}

func New(dyn dynamic.Interface, gatewayName, gatewayNamespace string) *Client {
	return &Client{
		dyn:              dyn,
		gatewayName:      gatewayName,
		gatewayNamespace: gatewayNamespace,
	}
}

var gatewayGVR = schema.GroupVersionResource{
	Group:    "gateway.networking.k8s.io",
	Version:  "v1",
	Resource: "gateways",
}

const listenerPrefix = "custom-"

func resourceNameFor(hostname string) string {
	return listenerPrefix + strings.ReplaceAll(hostname, ".", "-")
}

func isManagedListener(name string) bool {
	return strings.HasPrefix(name, listenerPrefix) && name != "custom-http" && name != "custom-https"
}
