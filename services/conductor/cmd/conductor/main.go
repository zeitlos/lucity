// Command conductor is the unified Lucity control-plane binary.
//
// It absorbs what were previously the gateway, packager, deployer,
// builder, and webhook services. The actual wiring lives in later
// phases of the conductor refactor; this entry point is currently a
// placeholder so the module compiles end-to-end.
package main

import (
	"log/slog"
	"os"
)

func main() {
	slog.Info("conductor: skeleton placeholder — wiring lands in phase 3")
	os.Exit(0)
}
