package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/zeitlos/lucity/cli/plugin"
)

const claudeUsage = `lucity claude — set up Claude Code

Usage:
  lucity claude plugin

Commands:
  plugin   Write the Lucity plugin for Claude Code to a local directory and
           print its path

The Lucity plugin marketplace runs 'lucity claude plugin' to install the plugin,
so the deploy skill always comes from the same version as the MCP server. Update
the CLI, and Claude Code picks up the matching plugin in its next session.

The plugin lands in the 'lucity/claude-plugin' directory of your user cache
directory, and older versions there are removed.
`

func cmdClaude(args []string) error {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, claudeUsage)
		return errors.New("missing claude subcommand")
	}
	switch args[0] {
	case "plugin":
		return claudePlugin(args[1:])
	case "help", "--help", "-h":
		fmt.Print(claudeUsage)
		return nil
	default:
		fmt.Fprintf(os.Stderr, "unknown claude subcommand %q\n\n%s", args[0], claudeUsage)
		return errors.New("unknown claude subcommand")
	}
}

func claudePlugin(args []string) error {
	flags := flag.NewFlagSet("claude plugin", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, claudeUsage) }
	if err := parseFlags(flags, args); err != nil {
		return err
	}

	cacheDirectory, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	directory, err := writePlugin(filepath.Join(cacheDirectory, "lucity", "claude-plugin"), plugin.Files)
	if err != nil {
		return err
	}
	fmt.Println(directory)
	return nil
}

func writePlugin(root string, files fs.FS) (string, error) {
	digest, err := pluginDigest(files)
	if err != nil {
		return "", err
	}
	directory := filepath.Join(root, digest)

	if _, err := os.Stat(directory); errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(root, 0o755); err != nil {
			return "", err
		}
		staging, err := os.MkdirTemp(root, ".staging-")
		if err != nil {
			return "", err
		}
		defer os.RemoveAll(staging)
		if err := os.CopyFS(staging, files); err != nil {
			return "", err
		}
		if err := os.Rename(staging, directory); err != nil {
			if _, statErr := os.Stat(directory); statErr != nil {
				return "", err
			}
		}
	} else if err != nil {
		return "", err
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.Name() == digest || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if err := os.RemoveAll(filepath.Join(root, entry.Name())); err != nil {
			return "", err
		}
	}
	return directory, nil
}

func pluginDigest(files fs.FS) (string, error) {
	hash := sha256.New()
	err := fs.WalkDir(files, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		file, err := files.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		fmt.Fprintf(hash, "%s\x00", path)
		if _, err := io.Copy(hash, file); err != nil {
			return err
		}
		hash.Write([]byte{0})
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil))[:12], nil
}
