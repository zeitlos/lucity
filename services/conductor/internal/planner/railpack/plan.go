package railpack

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/logger"
	"github.com/zeitlos/lucity/services/conductor/internal/planner"
)

func (c *Client) Plan(ctx context.Context, repoURL, ref, token string) ([]planner.Plan, error) {
	tmpDir, err := os.MkdirTemp("", "build-*")

	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(tmpDir)

	repo, err := git.PlainInit(tmpDir, false)

	if err != nil {
		return nil, err
	}

	if _, err := repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	}); err != nil {
		return nil, err
	}

	auth := &githttp.BasicAuth{Username: "x-access-token", Password: token}

	if err := repo.FetchContext(ctx, &git.FetchOptions{
		Auth:  auth,
		Depth: 1,
		RefSpecs: []config.RefSpec{
			config.RefSpec(ref + ":refs/heads/lucity-plan"),
		},
	}); err != nil {
		return nil, fmt.Errorf("fetch %s: %w", ref, err)
	}

	worktree, err := repo.Worktree()

	if err != nil {
		return nil, err
	}

	if err := worktree.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(ref),
	}); err != nil {
		return nil, fmt.Errorf("checkout %s: %w", ref, err)
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
