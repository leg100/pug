package tui

//go:generate stringer -type=Kind

type Kind int

const (
	ModuleListKind Kind = iota
	WorkspaceListKind
	StateKind
	RunListKind
	RunKind
	TaskListKind
	TaskKind
	LogListKind
	LogKind
)
