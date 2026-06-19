package values

// TODO: Move these labels and annotations to a shared pkg.
const (
	labelWorkspace          = "lucity.dev/workspace"
	labelProject            = "lucity.dev/project"
	labelEnvironment        = "lucity.dev/environment"
	labelService            = "lucity.dev/service"
	labelDatabase           = "lucity.dev/database"
	labelGitHubInstallation = "lucity.dev/github-installation"

	annotationSourceRepo    = "lucity.dev/source-repo"
	annotationSourceBranch  = "lucity.dev/source-branch"
	annotationSourceContext = "lucity.dev/source-context"
	annotationSourceMessage = "lucity.dev/source-commit-message"
)

func CommonLabels(workspace, project, environment string) map[string]string {
	return map[string]string{
		labelWorkspace:   workspace,
		labelProject:     project,
		labelEnvironment: environment,
	}
}
