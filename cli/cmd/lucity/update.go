package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"golang.org/x/term"

	"github.com/zeitlos/lucity/cli/internal/selfupdate"
)

const updateUsage = `lucity update — update the CLI to the latest release

Usage:
  lucity update

Downloads the latest release from GitHub, checks it against the release
signature, and replaces this binary with it.

The CLI also does this on its own: once a day it looks for a newer release, and
updates itself in the background while you keep working. Set
LUCITY_NO_AUTO_UPDATE to only be told about new releases instead.

When the binary sits in a directory only root can write to, such as
'/usr/local/bin', it can't update itself. Run 'sudo lucity update' there.
`

func cmdUpdate(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("update", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, updateUsage) }
	background := flags.Bool("background", false, "")
	if err := parseFlags(flags, args); err != nil {
		return err
	}
	if *background {
		backgroundUpdate(ctx)
		return nil
	}

	unlock, locked, err := selfupdate.Lock()
	if err != nil {
		return err
	}
	if !locked {
		return errors.New("another update is already running")
	}
	defer unlock()

	latest, updated, err := selfupdate.Update(ctx, version)
	if err != nil {
		return err
	}
	if !updated {
		fmt.Fprintf(os.Stderr, "lucity %s is up to date\n", version)
		return nil
	}
	fmt.Fprintf(os.Stderr, "updated lucity from %s to %s\n", version, latest)
	return nil
}

func backgroundUpdate(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	unlock, locked, err := selfupdate.Lock()
	if err != nil || !locked {
		return
	}
	defer unlock()
	if _, stale := selfupdate.Cached(); !stale {
		return
	}
	latest, updated, err := selfupdate.Update(ctx, version)
	if err == nil && updated {
		selfupdate.MarkUpdated(version, latest)
	}
}

func startUpdateCheck(ctx context.Context, command string) <-chan string {
	result := make(chan string, 1)
	if command == "mcp" || command == "claude" || command == "update" || selfupdate.Disabled() || version == "dev" {
		close(result)
		return result
	}
	if selfupdate.AutoUpdates(version) {
		latest, stale := selfupdate.Cached()
		if stale {
			selfupdate.StartBackground()
		}
		result <- latest
		return result
	}
	if os.Getenv("CI") != "" || !stderrIsTerminal() {
		close(result)
		return result
	}
	go func() {
		if command == "version" || command == "--version" || command == "-v" {
			latest, _ := selfupdate.Refresh(ctx)
			result <- latest
			return
		}
		result <- selfupdate.Latest(ctx)
	}()
	return result
}

func printUpdateNotice(result <-chan string) {
	latest, open := <-result
	if !open || !stderrIsTerminal() {
		return
	}
	if from, ok := selfupdate.Updated(version); ok {
		fmt.Fprintf(os.Stderr, "\nlucity updated itself from %s to %s\n", from, version)
	}
	if notice := selfupdate.Notice(version, latest); notice != "" {
		fmt.Fprintf(os.Stderr, "\n%s\n", notice)
	}
}

func stderrIsTerminal() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}
