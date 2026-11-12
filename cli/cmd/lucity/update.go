package main

import (
	"context"
	"fmt"
	"os"

	"github.com/zeitlos/lucity/cli/internal/updatecheck"
)

func startUpdateCheck(ctx context.Context, command string) <-chan string {
	result := make(chan string, 1)
	if command == "mcp" || !updateNoticeWanted() {
		close(result)
		return result
	}
	go func() {
		if command == "version" || command == "--version" || command == "-v" {
			latest, _ := updatecheck.Refresh(ctx)
			result <- latest
			return
		}
		result <- updatecheck.Latest(ctx)
	}()
	return result
}

func printUpdateNotice(result <-chan string) {
	latest := <-result
	if !updatecheck.Newer(version, latest) {
		return
	}
	fmt.Fprintf(os.Stderr, "\nlucity %s is available (you have %s): %s\n", latest, version, updatecheck.UpgradeHint())
}

func updateNoticeWanted() bool {
	if updatecheck.Disabled() || os.Getenv("CI") != "" || version == "dev" {
		return false
	}
	info, err := os.Stderr.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
