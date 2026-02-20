package charts

import "embed"

// LucityApp embeds the lucity-app Helm chart files.
// This is imported by the packager to inline the chart into GitOps repos.
//
//go:embed all:lucity-app
var LucityApp embed.FS
