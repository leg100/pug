package tui

import (
	"fmt"

	"golang.org/x/exp/maps"
)

type Kind int

const (
	ModuleListKind Kind = iota
	WorkspaceListKind
	RunListKind
	TaskListKind
	RunKind
	TaskKind
	LogsKind
)

var firstPages = map[string]Kind{
	"modules":    ModuleListKind,
	"workspaces": WorkspaceListKind,
	"runs":       RunListKind,
	"tasks":      TaskListKind,
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
