// Command builder is the in-pod build runner that the conductor's
// Build Jobs invoke. It clones the user's source repo, generates a
// railpack plan, executes the plan against BuildKit, pushes the
// resulting image to the platform's OCI registry, and annotates the
// parent K8s Job with the result.
//
// This binary has no API surface — it reads its inputs from
// environment variables (set by the conductor when it constructs the
// Job spec) and signals back via Job annotations on
// `lucity.dev/result` (success: image ref + digest) or
// `lucity.dev/error` (failure: free-text message). The conductor's
// build orchestrator (services/conductor/internal/builds) reads those
// annotations to surface the build outcome.
//
// The binary runs to completion and exits.
package main

import (
	"github.com/zeitlos/lucity/pkg/logger"
)

func main() {
	logger.Setup("info")
	runBuild()
}
