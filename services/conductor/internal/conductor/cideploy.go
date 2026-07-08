package conductor

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

const githubActionsSubjectPrefix = "github-actions:"

func GitHubActionsSubject(repository string) string {
	return githubActionsSubjectPrefix + repository
}

func GitHubActionsRepo(subject string) (string, bool) {
	return strings.CutPrefix(subject, githubActionsSubjectPrefix)
}

type CIDeployMatch struct {
	Services   []platform.ServiceID
	Workspaces []string
}

func (c *Client) MatchCIDeploy(ctx context.Context, repository, ref string) (*CIDeployMatch, error) {
	repoURL := "https://github.com/" + repository

	candidates, err := c.platform.ServicesByRepo(ctx, repoURL)

	if err != nil {
		return nil, err
	}

	var enabled []platform.RepoService

	for _, service := range candidates {
		if service.CIDeploy {
			enabled = append(enabled, service)
		}
	}

	if len(enabled) == 0 {
		return nil, fmt.Errorf("no service connected to %s has CI deploys enabled — enable it in the service settings", repository)
	}

	installationID, err := c.gitHubApp.FindInstallation(ctx, repository)

	if err != nil {
		return nil, fmt.Errorf("the Lucity GitHub App is not installed on %s", repository)
	}

	match := &CIDeployMatch{}
	seen := map[string]bool{}

	for _, service := range enabled {
		if service.InstallationID != 0 && service.InstallationID != installationID {
			continue
		}

		if !ciRefAllowed(ref, service.Branch) {
			continue
		}

		match.Services = append(match.Services, service.ID)

		if !seen[service.ID.Workspace] {
			seen[service.ID.Workspace] = true
			match.Workspaces = append(match.Workspaces, service.ID.Workspace)
		}
	}

	if len(match.Services) == 0 {
		return nil, fmt.Errorf("ref %q is not the tracked branch of any CI-deploy service connected to %s", ref, repository)
	}

	return match, nil
}

func ciRefAllowed(ref, branch string) bool {
	if branch != "" {
		return ref == "refs/heads/"+branch
	}

	return strings.HasPrefix(ref, "refs/heads/")
}
