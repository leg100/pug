package resource

type Kind int

const (
	Module Kind = iota
	Workspace
	Run
	Task
)
