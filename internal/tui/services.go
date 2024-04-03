package tui

import (
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type ModuleService interface {
	Get(id resource.ID) (*module.Module, error)
	List() []*module.Module
	Reload() error
	Init(moduleID resource.ID) (*task.Task, error)
	Format(moduleID resource.ID) (*task.Task, error)
	Validate(moduleID resource.ID) (*task.Task, error)
	SetCurrent(moduleID resource.ID, workspace resource.ID) error
}

type WorkspaceService interface {
	ReloadAll() (task.Multi, []error)
	Get(id resource.ID) (*workspace.Workspace, error)
	List(opts workspace.ListOptions) []*workspace.Workspace
}

type StateService interface {
	Reload(workspaceID resource.ID) (*task.Task, error)
	Get(workspaceID resource.ID) (*state.State, error)
	Delete(workspaceID resource.ID, addrs ...state.ResourceAddress) (*task.Task, error)
}

type RunService interface {
	Create(workspaceID resource.ID, opts run.CreateOptions) (*run.Run, error)
	Get(id resource.ID) (*run.Run, error)
	List(opts run.ListOptions) []*run.Run
	Apply(runID resource.ID) (*task.Task, error)
}

type TaskService interface {
	Counter() int
	Get(taskID resource.ID) (*task.Task, error)
	List(opts task.ListOptions) []*task.Task
	Cancel(taskID resource.ID) (*task.Task, error)
}
