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

const (
	resourcePrefix      = "custom-"
	resourceNameHashLen = 10
	resourceNameMaxLen  = 200
)

func ResourceNameFor(hostname string) string {
	hostname = strings.ToLower(hostname)
	name := strings.ReplaceAll(hostname, ".", "-")

	if len(name) > resourceNameMaxLen {
		name = name[:resourceNameMaxLen]
	}

	// Suffixing with hash of the hostname to avoid name collisions (foo-bar.com vs foo.bar.com) which could result in a takeover.
	sum := sha256.Sum256([]byte(hostname))

	return resourcePrefix + name + "-" + hex.EncodeToString(sum[:])[:resourceNameHashLen]
}

func isManagedListener(name string) bool {
	return strings.HasPrefix(name, resourcePrefix) && name != "custom-http" && name != "custom-https"
}
