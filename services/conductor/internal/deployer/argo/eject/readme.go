package eject

import (
	"fmt"
	"strings"
)

// readmeContent generates the README.md for an ejected project.
func readmeContent(project string, environments []string, services []serviceInfo) string {
	var b strings.Builder

	shortName := project
	if parts := strings.SplitN(project, "/", 2); len(parts) == 2 {
		shortName = parts[1]
	}

	b.WriteString(fmt.Sprintf(`# %s — Ejected from Lucity

This directory contains a fully self-contained deployment configuration
ejected from [Lucity](https://lucity.cloud). You can run it independently
with standard Kubernetes, Helm, and ArgoCD tooling. No Lucity dependency remains.

## Prerequisites

- Kubernetes cluster (1.27+)
- [Helm](https://helm.sh/) 3.x
- [ArgoCD](https://argo-cd.readthedocs.io/) installed on the cluster
- [kubectl](https://kubernetes.io/docs/tasks/tools/) configured for your cluster
- [railpack](https://docs.railway.com/railpack) CLI (for building images)
- A container registry (Docker Hub, GHCR, ECR, etc.)

## Directory Structure

`+"```"+`
%s-ejected/
  base/
    Chart.yaml              Helm chart — depends on the lucity-app subchart
    values.yaml             Shared service definitions
  environments/
`, shortName, shortName))

	for _, env := range environments {
		b.WriteString(fmt.Sprintf("    %s/values.yaml%s\n", env, envDescription(env)))
	}

	b.WriteString(`  chart/
    Chart.yaml              Complete Helm chart (templates + defaults)
    values.yaml
    templates/              Kubernetes resource templates
  argocd/
    applications/           One ArgoCD Application per environment
  build.sh                  Build script using railpack
  README.md                 This file
` + "```" + `

## Quick Start

### 1. Push to your Git repo

Create a new Git repository and push this directory:

` + "```" + `bash
cd ` + shortName + `-ejected
git init
git add .
git commit -m "Initial ejected configuration"
git remote add origin <YOUR_REPO_URL>
git push -u origin main
` + "```" + `

### 2. Update ArgoCD manifests

Edit each file in ` + "`argocd/applications/`" + ` and replace
` + "`<YOUR_GITOPS_REPO_URL>`" + ` with your actual Git repo URL.

### 3. Build and push images

`)

	if len(services) > 0 {
		b.WriteString("```bash\n")
		b.WriteString(fmt.Sprintf("REGISTRY=ghcr.io/your-org TAG=$(git rev-parse --short HEAD) ./build.sh\n"))
		b.WriteString("```\n\n")

		b.WriteString("This builds and pushes images for:\n\n")
		for _, svc := range services {
			b.WriteString(fmt.Sprintf("- **%s** (`%s`)\n", svc.Name, svc.Image))
		}
		b.WriteString("\n")
	} else {
		b.WriteString("Configure your services in `base/values.yaml` first, then use `build.sh`.\n\n")
	}

	b.WriteString(`### 4. Update image tags

After building, update the image tags in each environment's values file.
For example, in ` + "`environments/development/values.yaml`" + `:

` + "```" + `yaml
lucity-app:
  services:
    web:
      image:
        tag: "abc1234"
` + "```" + `

Commit and push — ArgoCD will sync automatically.

### 5. Apply ArgoCD Applications

` + "```" + `bash
kubectl apply -f argocd/applications/
` + "```" + `

ArgoCD will create the namespaces and deploy your services.

## Environment Promotion

To promote an image from one environment to another, copy the image tag:

1. Check the current tag in ` + "`environments/development/values.yaml`" + `
2. Set the same tag in ` + "`environments/staging/values.yaml`" + `
3. Commit and push — ArgoCD syncs automatically

## Modifying Services

Service definitions live in ` + "`base/values.yaml`" + ` under the ` + "`lucity-app:`" + ` key.
You can add, remove, or modify services there. See ` + "`chart/values.yaml`" + ` for
all available options (replicas, resources, environment variables, etc.).

## Adding an Environment

1. Create ` + "`environments/<name>/values.yaml`" + ` with your overrides
2. Create an ArgoCD Application manifest in ` + "`argocd/applications/`" + `
3. Apply: ` + "`kubectl apply -f argocd/applications/<name>.yaml`" + `

---

*Ejected from Lucity. Your infrastructure, your rules.*
`)

	return b.String()
}

func envDescription(env string) string {
	switch env {
	case "development":
		return "  Development overrides"
	case "staging":
		return "      Staging overrides"
	case "production":
		return "     Production overrides"
	default:
		return ""
	}
}
