package tui

import (
	"context"

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
	Reload(ctx context.Context) chan module.ReloadResult
	Init(moduleID resource.ID) (task.Spec, error)
	Format(moduleID resource.ID) (task.Spec, error)
	Validate(moduleID resource.ID) (task.Spec, error)
	SetCurrent(moduleID, workspaceID resource.ID) error
	Subscribe() <-chan resource.Event[*module.Module]
	Shutdown()
}

type WorkspaceService interface {
	Reload(moduleID resource.ID) (task.Spec, error)
	Get(id resource.ID) (*workspace.Workspace, error)
	List(opts workspace.ListOptions) []*workspace.Workspace
	SelectWorkspace(moduleID, workspaceID resource.ID) error
	Delete(id resource.ID) (task.Spec, error)
	Subscribe() <-chan resource.Event[*workspace.Workspace]
	Shutdown()
}

type StateService interface {
	Reload(workspaceID resource.ID) (task.Spec, error)
	Get(workspaceID resource.ID) (*state.State, error)
	GetResource(resourceID resource.ID) (*state.Resource, error)
	Delete(workspaceID resource.ID, addrs ...state.ResourceAddress) (task.Spec, error)
	Taint(workspaceID resource.ID, addr state.ResourceAddress) (task.Spec, error)
	Untaint(workspaceID resource.ID, addr state.ResourceAddress) (task.Spec, error)
	Move(workspaceID resource.ID, src, dest state.ResourceAddress) (task.Spec, error)
	Subscribe() <-chan resource.Event[*state.State]
	Shutdown()
}

type RunService interface {
	Get(id resource.ID) (*run.Run, error)
	List(opts run.ListOptions) []*run.Run
	Plan(workspaceID resource.ID, opts run.CreateOptions) (task.Spec, error)
	Apply(id resource.ID, opts *run.CreateOptions) (task.Spec, error)
	Subscribe() <-chan resource.Event[*run.Run]
	Shutdown()
}

type TaskService interface {
	Create(opts task.Spec) (*task.Task, error)
	CreateGroup(specs ...task.Spec) (*task.Group, error)
	Counter() int
	Get(taskID resource.ID) (*task.Task, error)
	GetGroup(groupID resource.ID) (*task.Group, error)
	List(opts task.ListOptions) []*task.Task
	ListGroups() []*task.Group
	Cancel(taskID resource.ID) (*task.Task, error)
	Subscribe() <-chan resource.Event[*task.Task]
	SubscribeGroups() <-chan resource.Event[*task.Group]
	Shutdown()
	ShutdownGroups()
}
