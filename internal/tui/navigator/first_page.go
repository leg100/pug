package navigator

import (
	"fmt"

	"github.com/leg100/pug/internal/tui"
	"golang.org/x/exp/maps"
)

var firstPages = map[string]tui.Kind{
	"modules":    tui.ModuleListKind,
	"workspaces": tui.WorkspaceListKind,
	"runs":       tui.RunListKind,
	"tasks":      tui.TaskListKind,
	"logs":       tui.LogListKind,
}

// firstPageKind retrieves the model corresponding to the user requested first
// page.
func firstPageKind(s string) (tui.Kind, error) {
	kind, ok := firstPages[s]
	if !ok {
		return 0, fmt.Errorf("invalid first page, must be one of: %v", maps.Keys(firstPages))
	}
	return kind, nil
}
