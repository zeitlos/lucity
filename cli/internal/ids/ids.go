package ids

import (
	"errors"
	"fmt"
	"strings"
)

// Scoped validates and normalizes a composite Lucity id to segments
// '/'-separated parts. A workspace-relative id (segments-1 parts) is expanded by
// prefixing the active workspace; a fully-qualified id must name that workspace.
func Scoped(workspace, id string, segments int) (string, error) {
	id = strings.Trim(strings.TrimSpace(id), "/")
	if id == "" {
		return "", errors.New("id must not be empty")
	}

	parts := strings.Split(id, "/")

	switch len(parts) {
	case segments - 1:
		return workspace + "/" + id, nil
	case segments:
		if parts[0] != workspace {
			return "", fmt.Errorf("id belongs to workspace %s but your active workspace is %s — run 'lucity workspace %s'", parts[0], workspace, parts[0])
		}
		return id, nil
	default:
		return "", fmt.Errorf("invalid id %q: expected %d '/'-separated segments (or %d relative to the active workspace)", id, segments, segments-1)
	}
}

func Project(workspace, id string) (string, error)     { return Scoped(workspace, id, 2) }
func Environment(workspace, id string) (string, error) { return Scoped(workspace, id, 3) }
func Service(workspace, id string) (string, error)     { return Scoped(workspace, id, 4) }
func Resource(workspace, id string) (string, error)    { return Scoped(workspace, id, 4) }
