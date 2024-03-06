package tui

import "github.com/leg100/pug/internal/resource"

type modelKind int

const (
	ModuleListKind modelKind = iota
	WorkspaceListKind
	RunListKind
	TaskListKind
	TaskKind
	LogsKind
	HelpKind
)

// page identifies an instance of a model
type page struct {
	kind     modelKind
	resource resource.Resource
}

func (p page) cacheKey() cacheKey {
	return cacheKey{kind: p.kind, id: p.resource.ID}
}
