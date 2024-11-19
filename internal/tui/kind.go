package tui

//go:generate stringer -type=Kind

type Kind int

const (
	ModuleListKind Kind = iota
	WorkspaceListKind
	TaskListKind
	TaskKind
	TaskGroupListKind
	TaskGroupKind
	ResourceListKind
	ResourceKind
	LogListKind
	LogKind
	ExplorerKind
)
