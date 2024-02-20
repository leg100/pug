package workspace

import (
	"fmt"

	"github.com/leg100/pug/internal/task"
)

// ID uniquely identifies a workspace
type ID struct {
	Path string
	Name string
}

func (id ID) String() string {
	return fmt.Sprintf("%s:%s", id.Path, id.Name)
}

func (id ID) ID() string {
	return id.String()
}

type Workspace struct {
	ID
}

const (
	PlanTask    task.Kind = "plan"
	ApplyTask   task.Kind = "apply"
	RefreshTask task.Kind = "refresh"
	DestroyTask task.Kind = "destroy"
	TaintTask   task.Kind = "taint"
)
