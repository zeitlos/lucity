// Package webhook handles incoming GitHub webhook events: signature
// verification, event parsing, and dispatch into the build/deploy
// pipeline.
//
// After the conductor merge, the pipeline calls handler functions
// directly rather than crossing gRPC to packager + deployer + builder.
//
// Populated in phase 4 from services/webhook/{http,github,pipeline.go}.
package webhook
