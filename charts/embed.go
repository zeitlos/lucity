package charts

import "embed"

// LucityApp embeds the lucity-app Helm chart files.
//
//go:embed all:lucity-app
var LucityApp embed.FS
