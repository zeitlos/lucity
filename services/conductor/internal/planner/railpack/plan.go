package railpack

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/logger"

	"github.com/zeitlos/lucity/pkg/sourcetree"
	"github.com/zeitlos/lucity/pkg/usererror"
	"github.com/zeitlos/lucity/services/conductor/internal/planner"
)

func (c *Client) Plan(ctx context.Context, repoURL, ref, token, contextPath string) ([]planner.Plan, error) {
	tmpDir, err := os.MkdirTemp("", "plan-*")

	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(tmpDir)

	if err := shallowFetch(ctx, tmpDir, repoURL, ref, token); err != nil {
		return nil, err
	}

	if _, err := sourcetree.HideEscapingSymlinks(tmpDir); err != nil {
		return nil, fmt.Errorf("hide symlinks: %w", err)
	}

	dir, err := contextDir(tmpDir, contextPath)

	if err != nil {
		return nil, err
	}

	errUndetectable := usererror.New("couldn't detect how to build this repository: no supported language or framework found at its root or in its folders")

	if dir != "." {
		errUndetectable = usererror.Errorf("couldn't detect how to build %q: no supported language or framework found there or in its folders", dir)
	}

	repositoryName := strings.TrimSuffix(path.Base(repoURL), ".git")
	service, err := detect(tmpDir, dir, repositoryName)

	if err == nil && service.StartCommand != "" {
		return []planner.Plan{service}, nil
	}

	folders, scanErr := scan(ctx, tmpDir, dir)

	if scanErr != nil {
		return nil, scanErr
	}

	if len(folders) > 0 {
		return folders, nil
	}

	if err != nil {
		slog.WarnContext(ctx, "service detection failed", "repository", repoURL, "error", err)
		return nil, errUndetectable
	}

	return []planner.Plan{service}, nil
}

const (
	folderDepth = 2
	folderLimit = 100
)

func scan(ctx context.Context, repoPath, dir string) ([]planner.Plan, error) {
	var services []planner.Plan

	parents := []string{dir}
	scanned := 0

	for range folderDepth {
		var next []string

		for _, parent := range parents {
			entries, err := os.ReadDir(filepath.Join(repoPath, parent))

			if err != nil {
				continue
			}

			for _, entry := range entries {
				if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
					continue
				}

				if scanned == folderLimit {
					return services, nil
				}

				if err := ctx.Err(); err != nil {
					return nil, err
				}

				scanned++
				folder := filepath.Join(parent, entry.Name())
				service, err := detect(repoPath, folder, "")

				if err != nil {
					next = append(next, folder)
					continue
				}

				services = append(services, service)
			}
		}

		parents = next
	}

	return services, nil
}

func detect(repoPath, dir, repositoryName string) (planner.Plan, error) {
	a, err := app.NewApp(filepath.Join(repoPath, dir))

	if err != nil {
		return planner.Plan{}, err
	}

	generated, err := core.GenerateBuildPlan(a, app.NewEnvironment(nil), &core.GenerateBuildPlanOptions{})

	if err != nil {
		return planner.Plan{}, err
	}

	if !generated.Success || generated.Plan == nil {
		var errs []string

		for _, l := range generated.Logs {
			if l.Level == logger.Error {
				errs = append(errs, l.Msg)
			}
		}

		if len(errs) == 0 {
			return planner.Plan{}, errors.New("unknown error")
		}

		return planner.Plan{}, errors.New(strings.Join(errs, "; "))
	}

	service := planner.Plan{
		Name:         repositoryName,
		Providers:    generated.DetectedProviders,
		StartCommand: generated.Plan.Deploy.StartCmd,
		ContextPath:  "/",
		Versions:     make(map[string]string, len(generated.ResolvedPackages)),
	}

	if dir != "." {
		service.Name = filepath.Base(dir)
		service.ContextPath = filepath.ToSlash(dir)
	}

	for name, details := range generated.ResolvedPackages {
		if details.ResolvedVersion == nil {
			continue
		}

		service.Versions[name] = *details.ResolvedVersion
	}

	return service, nil
}

func contextDir(repoPath, contextPath string) (string, error) {
	root, err := os.OpenRoot(repoPath)

	if err != nil {
		return "", err
	}

	defer root.Close()

	dir := filepath.Join(".", contextPath)
	info, err := root.Stat(dir)

	if err != nil || !info.IsDir() {
		return "", usererror.Errorf("%q is not a folder of this repository", contextPath)
	}

	return dir, nil
}

func shallowFetch(ctx context.Context, dir, repoURL, ref, token string) error {
	env := append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	if token != "" {
		header := "Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte("x-access-token:"+token))
		env = append(env,
			"GIT_CONFIG_COUNT=1",
			"GIT_CONFIG_KEY_0=http.extraHeader",
			"GIT_CONFIG_VALUE_0="+header,
		)
	}

	run := func(args ...string) error {
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = dir
		cmd.Env = env
		cmd.Stderr = os.Stderr

		return cmd.Run()
	}

	if err := run("init", "--quiet"); err != nil {
		return fmt.Errorf("git init: %w", err)
	}

	if err := run("remote", "add", "origin", repoURL); err != nil {
		return fmt.Errorf("git remote add: %w", err)
	}

	if err := run("fetch", "--quiet", "--depth", "1", "origin", ref); err != nil {
		return fmt.Errorf("git fetch %s: %w", ref, err)
	}

	if err := run("checkout", "--quiet", "FETCH_HEAD"); err != nil {
		return fmt.Errorf("git checkout %s: %w", ref, err)
	}

	return nil
}
