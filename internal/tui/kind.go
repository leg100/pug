package tui

//go:generate stringer -type=Kind

type Kind int

const (
	TaskListKind Kind = iota
	TaskKind
	TaskGroupListKind
	TaskGroupKind
	ResourceListKind
	ResourceKind
	LogListKind
	LogKind
	ExplorerKind
)
