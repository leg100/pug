package tui

import (
	"fmt"

	"golang.org/x/exp/maps"
)

var firstPages = map[string]Kind{
	"modules":    ModuleListKind,
	"workspaces": WorkspaceListKind,
	"tasks":      TaskListKind,
	"logs":       LogListKind,
}

// FirstPageKind retrieves the model corresponding to the user requested first
// page.
func FirstPageKind(s string) (Kind, error) {
	kind, ok := firstPages[s]
	if !ok {
		return 0, fmt.Errorf("invalid first page, must be one of: %v", maps.Keys(firstPages))
	}
	return kind, nil
}
