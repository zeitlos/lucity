package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Client struct {
	dyn              dynamic.Interface
	gatewayName      string
	gatewayNamespace string
	clusterIssuer    string
}

func New(dyn dynamic.Interface, gatewayName, gatewayNamespace, clusterIssuer string) *Client {
	return &Client{
		dyn:              dyn,
		gatewayName:      gatewayName,
		gatewayNamespace: gatewayNamespace,
		clusterIssuer:    clusterIssuer,
	}
}

var gatewayGVR = schema.GroupVersionResource{
	Group:    "gateway.networking.k8s.io",
	Version:  "v1",
	Resource: "gateways",
}

var certificateGVR = schema.GroupVersionResource{
	Group:    "cert-manager.io",
	Version:  "v1",
	Resource: "certificates",
}

var secretGVR = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "secrets",
}

const listenerPrefix = "custom-"

func ResourceNameFor(hostname string) string {
	hostname = strings.ToLower(hostname)
	sum := sha256.Sum256([]byte(hostname))
	name := strings.ReplaceAll(hostname, ".", "-")

	if len(name) > 200 {
		name = name[:200]
	}

	return listenerPrefix + name + "-" + hex.EncodeToString(sum[:4])
}

func isManagedListener(name string) bool {
	return strings.HasPrefix(name, listenerPrefix) && name != "custom-http" && name != "custom-https"
}
