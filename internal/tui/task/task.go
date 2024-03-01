package task

import (
	"github.com/leg100/pug/internal/module"
	taskpkg "github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type factory struct {
	modules    *module.Service
	workspaces *workspace.Service
}

func (f *factory) newTask(upstream *taskpkg.Task) *task {
}

// task wraps the upstream task, adding info about the task that is useful to
// the user, like its module path, workspace name, etc.
type task struct {
}
