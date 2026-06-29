package values

// TODO: Move these labels and annotations to a shared pkg.
const (
	labelWorkspace          = "lucity.dev/workspace"
	labelProject            = "lucity.dev/project"
	labelEnvironment        = "lucity.dev/environment"
	labelService            = "lucity.dev/service"
	labelDatabase           = "lucity.dev/database"
	labelKeyValueStore      = "lucity.dev/keyvaluestore"
	labelVolume             = "lucity.dev/volume"
	labelSharedVariables    = "lucity.dev/shared-variables"
	labelGitHubInstallation = "lucity.dev/github-installation"

	annotationDatabaseHost  = "lucity.dev/db-host"
	annotationSourceRepo    = "lucity.dev/source-repo"
	annotationSourceBranch  = "lucity.dev/source-branch"
	annotationSourceContext = "lucity.dev/source-context"
	annotationSourceCommit  = "lucity.dev/source-commit"
	annotationSourceMessage = "lucity.dev/source-commit-message"
	annotationBuildID       = "lucity.dev/build-id"

	annotationRelease        = "lucity.dev/release"
	annotationReleaseTrigger = "lucity.dev/release-trigger"
	annotationReleaseActor   = "lucity.dev/release-actor"
)

func CommonLabels(workspace, project, environment string) map[string]string {
	return map[string]string{
		labelWorkspace:   workspace,
		labelProject:     project,
		labelEnvironment: environment,
	}
}

func SharedVariableLabels() map[string]string {
	return map[string]string{labelSharedVariables: "true"}
}
