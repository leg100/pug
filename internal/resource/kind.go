package resource

import "fmt"

type Kind int

const (
	Global Kind = iota
	Module
	Workspace
	Run
	Task
	Log
)

func (k Kind) String() string {
	return [...]string{
		"global",
		"mod",
		"ws",
		"run",
		"task",
		"log",
	}[k]
}

var kindMap = map[string]Kind{
	"global": Global,
	"mod":    Module,
	"ws":     Workspace,
	"run":    Run,
	"task":   Task,
	"lod":    Log,
}

func kindString(s string) (Kind, error) {
	kind, ok := kindMap[s]
	if !ok {
		return 0, fmt.Errorf("cannot parse kind from string: %s", s)
	}
	return kind, nil
}
