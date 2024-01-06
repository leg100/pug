package resource

type Kind int

const (
	Global Kind = iota
	Module
	Workspace
	Run
	Task
	Log
	StateResource
)

func (k Kind) String() string {
	return [...]string{
		"global",
		"mod",
		"ws",
		"run",
		"task",
		"log",
		"res",
	}[k]
}
