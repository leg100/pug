package tui

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
