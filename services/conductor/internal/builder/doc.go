// Package builder orchestrates source-to-image builds for user
// workloads. It detects the source's language/framework via railpack,
// creates a Kubernetes Job that runs BuildKit, and tracks the build
// to completion.
//
// The build sandbox itself (BuildKit Deployment + Build Job pods)
// runs in the lucity-builds namespace and is unchanged by the
// conductor merge — only the orchestrator moves in-process.
//
// This package is populated in phase 2 by relocating
// services/builder/{engine,build}/.
package builder
