package helm

import (
	"bytes"
	"fmt"

	"helm.sh/helm/v4/pkg/postrenderer"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/yaml"
)

// releaseKinds are the kinds the lucity-app chart renders, all namespaced.
// A kind added to the chart must be added here, or every apply fails.
var releaseKinds = map[schema.GroupKind]bool{
	{Kind: "Secret"}:                                             true,
	{Kind: "Service"}:                                            true,
	{Kind: "PersistentVolumeClaim"}:                              true,
	{Group: "apps", Kind: "Deployment"}:                          true,
	{Group: "apps", Kind: "StatefulSet"}:                         true,
	{Group: "autoscaling", Kind: "HorizontalPodAutoscaler"}:      true,
	{Group: "autoscaling.k8s.io", Kind: "VerticalPodAutoscaler"}: true,
	{Group: "gateway.networking.k8s.io", Kind: "HTTPRoute"}:      true,
	{Group: "gateway.networking.k8s.io", Kind: "ListenerSet"}:    true,
	{Group: "cert-manager.io", Kind: "Certificate"}:              true,
	{Group: "postgresql.cnpg.io", Kind: "Cluster"}:               true,
	{Group: "postgresql.cnpg.io", Kind: "ScheduledBackup"}:       true,
	{Group: "barmancloud.cnpg.io", Kind: "ObjectStore"}:          true,
	{Group: "traefik.io", Kind: "IngressRouteTCP"}:               true,
	{Group: "traefik.io", Kind: "MiddlewareTCP"}:                 true,
}

// releaseGuard is a post-renderer that rejects a render containing a list,
// an object of a kind the chart does not render, or an object outside the
// release namespace. It decodes each document the way Helm does before
// applying it, and returns the manifests unchanged.
//
// It is a stopgap until tenant releases are applied by impersonating a
// namespace-scoped ServiceAccount, whose Role makes the apiserver enforce the
// same limits. Remove it once that lands.
type releaseGuard struct {
	namespace string
}

func (g releaseGuard) Run(manifests *bytes.Buffer) (*bytes.Buffer, error) {
	nodes, err := kio.ParseAll(manifests.String())

	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		doc, err := node.String()

		if err != nil {
			return nil, err
		}

		var object unstructured.Unstructured

		if err := yaml.Unmarshal([]byte(doc), &object.Object); err != nil {
			return nil, err
		}

		kind := object.GroupVersionKind().GroupKind()

		if object.IsList() || !releaseKinds[kind] {
			return nil, fmt.Errorf("release renders a disallowed object: %s %q", kind, object.GetName())
		}

		if namespace := object.GetNamespace(); namespace != "" && namespace != g.namespace {
			return nil, fmt.Errorf("release renders %s %q into namespace %q", kind, object.GetName(), namespace)
		}
	}

	return manifests, nil
}

var _ postrenderer.PostRenderer = releaseGuard{}
