package eject

import (
	"fmt"
	"strings"
)

// serviceInfo holds the minimal info needed for build script generation.
type serviceInfo struct {
	Name  string
	Image string
}

// buildScript generates a shell script for building project services with railpack.
func buildScript(services []serviceInfo) string {
	var b strings.Builder

	b.WriteString(`#!/usr/bin/env bash
set -euo pipefail

# Build script for your project services using railpack.
# Railpack detects your language/framework and builds an OCI image.
#
# Install railpack: https://docs.railway.com/railpack
#   npm install -g @aspect-build/railpack
#
# Usage:
#   ./build.sh                           # build all services with defaults
#   REGISTRY=ghcr.io/myorg TAG=v1.0 ./build.sh   # custom registry and tag

REGISTRY="${REGISTRY:-localhost:5000}"
TAG="${TAG:-latest}"

`)

	if len(services) == 0 {
		b.WriteString("echo \"No services configured.\"\n")
		return b.String()
	}

	for _, svc := range services {
		image := svc.Image
		// Replace any existing tag with the variable
		if idx := strings.LastIndex(image, ":"); idx != -1 {
			image = image[:idx]
		}

		b.WriteString(fmt.Sprintf(`echo "Building %s..."
railpack build --name "%s:${TAG}" .
docker push "%s:${TAG}"
echo ""

`, svc.Name, image, image))
	}

	b.WriteString("echo \"All services built and pushed.\"\n")
	return b.String()
}
