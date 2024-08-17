package resource

type Kind int

const (
	Global Kind = iota
	Module
	Workspace
	Plan
	Task
	TaskGroup
	Log
	LogAttr
	State
	StateResource
)

func (k Kind) String() string {
	return [...]string{
		"global",
		"mod",
		"ws",
		"run",
		"task",
		"tg",
		"log",
		"attr",
		"state",
		"res",
	}[k]
}
