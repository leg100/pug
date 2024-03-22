package tui

//go:generate stringer -type=Kind

type Kind int

const (
	ModuleListKind Kind = iota
	ModuleKind
	WorkspaceListKind
	WorkspaceKind
	RunListKind
	RunKind
	TaskListKind
	TaskKind
	LogsKind
)
