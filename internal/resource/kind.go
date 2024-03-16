package resource

type Kind int

const (
	Global Kind = iota
	Module
	Workspace
	Run
	Task
)
