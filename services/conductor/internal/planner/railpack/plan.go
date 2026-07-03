package railpack

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/logger"
	"github.com/zeitlos/lucity/services/conductor/internal/planner"
)

func (c *Client) Plan(ctx context.Context, repoURL, ref, token string) ([]planner.Plan, error) {
	tmpDir, err := os.MkdirTemp("", "plan-*")

	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(tmpDir)

	if err := shallowFetch(ctx, tmpDir, repoURL, ref, token); err != nil {
		return nil, err
	}

	a, err := app.NewApp(tmpDir)

	if err != nil {
		return nil, err
	}

	env := app.NewEnvironment(nil)
	generated := core.GenerateBuildPlan(a, env, &core.GenerateBuildPlanOptions{})

	if !generated.Success || generated.Plan == nil {
		errMsg := "unknown error"

		var errs []string
		for _, l := range generated.Logs {
			if l.Level == logger.Error {
				errs = append(errs, l.Msg)
			}
		}

		if len(errs) > 0 {
			errMsg = strings.Join(errs, "; ")
		}

		return nil, fmt.Errorf("service detection failed: %s", errMsg)
	}

	service := planner.Plan{
		Name:         path.Base(repoURL), // TODO: Verify if this works as expected.
		Providers:    generated.DetectedProviders,
		StartCommand: generated.Plan.Deploy.StartCmd,
		ContextPath:  "/",
		Versions:     make(map[string]string, len(generated.ResolvedPackages)),
	}

	for name, details := range generated.ResolvedPackages {
		if details.ResolvedVersion == nil {
			continue
		}

		service.Versions[name] = *details.ResolvedVersion
	}

	return []planner.Plan{service}, nil
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
