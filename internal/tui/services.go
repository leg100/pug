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
	Reload() ([]string, []string, error)
	Init(moduleID resource.ID) (*task.Task, error)
	Format(moduleID resource.ID) (*task.Task, error)
	Validate(moduleID resource.ID) (*task.Task, error)
	SetCurrent(moduleID, workspaceID resource.ID) error
}

type WorkspaceService interface {
	Reload(moduleID resource.ID) (*task.Task, error)
	Get(id resource.ID) (*workspace.Workspace, error)
	List(opts workspace.ListOptions) []*workspace.Workspace
	SelectWorkspace(moduleID, workspaceID resource.ID) error
	Delete(id resource.ID) (*task.Task, error)
}

type StateService interface {
	Reload(workspaceID resource.ID) (*task.Task, error)
	Get(workspaceID resource.ID) (*state.State, error)
	Delete(workspaceID resource.ID, addrs ...state.ResourceAddress) (*task.Task, error)
	Taint(workspaceID resource.ID, addr state.ResourceAddress) (*task.Task, error)
	Untaint(workspaceID resource.ID, addr state.ResourceAddress) (*task.Task, error)
	Move(workspaceID resource.ID, src, dest state.ResourceAddress) (*task.Task, error)
}

type RunService interface {
	Get(id resource.ID) (*run.Run, error)
	List(opts run.ListOptions) []*run.Run
	Plan(workspaceID resource.ID, opts run.CreateOptions) (*task.Task, error)
	Apply(opts *run.CreateOptions, ids ...resource.ID) (*task.Group, error)
}

type TaskService interface {
	CreateGroup(cmd string, fn task.Func, ids ...resource.ID) (*task.Group, error)
	Retry(taskID resource.ID) (*task.Task, error)
	Counter() int
	Get(taskID resource.ID) (*task.Task, error)
	List(opts task.ListOptions) []*task.Task
	ListGroups() []*task.Group
	Cancel(taskID resource.ID) (*task.Task, error)
}
