package signals

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

type Kind string

const (
	SharedCommit Kind = "shared-commit"
	Image        Kind = "image"
	Fork         Kind = "fork"
	YoungAccount Kind = "young-account"
)

type Signal struct {
	Kind   Kind
	Strong bool
	Detail string
}

// Workspace holds the facts about one workspace that signals are judged on.
type Workspace struct {
	ID          string
	CreatedAt   time.Time
	Commits     []string
	Images      []string
	ForkParents []string
	Accounts    []time.Time
}

const (
	sharedCommitWorkspaces = 3
	youngAccountAge        = 90 * 24 * time.Hour
)

var deniedImages = []string{
	"ghcr.io/omidiran-gaming/omid-3x-ui",
	"ghcr.io/mhsanaei/3x-ui",
}

var deniedForkParents = []string{
	"mhsanaei/3x-ui",
	"alireza0/x-ui",
	"vaxilu/x-ui",
	"xtls/xray-core",
}

// Evaluate returns the signals of every workspace, keyed by workspace ID.
func Evaluate(workspaces []Workspace) map[string][]Signal {
	sharing := map[string]map[string]bool{}

	for _, workspace := range workspaces {
		for _, commit := range workspace.Commits {
			if sharing[commit] == nil {
				sharing[commit] = map[string]bool{}
			}
			sharing[commit][workspace.ID] = true
		}
	}

	result := make(map[string][]Signal, len(workspaces))

	for _, workspace := range workspaces {
		var found []Signal

		for _, commit := range unique(workspace.Commits) {
			if count := len(sharing[commit]); count >= sharedCommitWorkspaces {
				found = append(found, Signal{Kind: SharedCommit, Strong: true, Detail: fmt.Sprintf("%s in %d workspaces", short(commit), count)})
			}
		}

		for _, image := range unique(workspace.Images) {
			if repository := imageRepository(image); slices.Contains(deniedImages, repository) {
				found = append(found, Signal{Kind: Image, Strong: true, Detail: repository})
			}
		}

		for _, parent := range unique(workspace.ForkParents) {
			if slices.Contains(deniedForkParents, strings.ToLower(parent)) {
				found = append(found, Signal{Kind: Fork, Strong: true, Detail: parent})
			}
		}

		for _, created := range workspace.Accounts {
			if age := workspace.CreatedAt.Sub(created); age < youngAccountAge {
				found = append(found, Signal{Kind: YoungAccount, Detail: fmt.Sprintf("%d days old", int(age.Hours()/24))})
				break
			}
		}

		result[workspace.ID] = found
	}

	return result
}

func imageRepository(image string) string {
	image, _, _ = strings.Cut(image, "@")

	if colon := strings.LastIndex(image, ":"); colon > strings.LastIndex(image, "/") {
		image = image[:colon]
	}

	return strings.ToLower(image)
}

func short(commit string) string {
	if len(commit) > 7 {
		return commit[:7]
	}
	return commit
}

func unique(values []string) []string {
	seen := map[string]bool{}
	var result []string

	for _, value := range values {
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}

	return result
}
