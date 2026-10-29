package charts

import "embed"

// LucityApp embeds the lucity-app Helm chart files.
//
//go:embed all:lucity-app
var LucityApp embed.FS

// LucityWorkspace embeds the lucity-workspace Helm chart files.
//
//go:embed all:lucity-workspace
var LucityWorkspace embed.FS
