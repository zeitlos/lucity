// Package plugin embeds the Lucity plugin for Claude Code, so the CLI ships the
// deploy skill at the same version as its MCP server.
package plugin

import "embed"

//go:embed .claude-plugin .mcp.json README.md skills
var Files embed.FS
