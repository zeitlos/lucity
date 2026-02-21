package eject

import (
	"fmt"

	"github.com/zeitlos/lucity/pkg/labels"
)

// applicationYAML generates an ArgoCD Application manifest for a single environment.
func applicationYAML(project, environment string) string {
	appName := labels.NamespaceFor(project, environment)

	return fmt.Sprintf(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: %s
  namespace: argocd
spec:
  source:
    repoURL: <YOUR_GITOPS_REPO_URL>
    path: base
    targetRevision: HEAD
    helm:
      valueFiles:
        - ../environments/%s/values.yaml
  destination:
    server: https://kubernetes.default.svc
    namespace: %s
  project: default
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
`, appName, environment, appName)
}
