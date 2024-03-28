package state

import (
	"context"
	"encoding/json"

	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type Service struct {
	workspaces *workspace.Service
	tasks      *task.Service
}

type ServiceOptions struct {
	WorkspaceService *workspace.Service
	TaskService      *task.Service
}

func NewService(ctx context.Context, opts ServiceOptions) *Service {
	svc := &Service{
		workspaces: opts.WorkspaceService,
		tasks:      opts.TaskService,
	}
	return svc
}

// Pull creates a task to pull the state for a workspace.
func (s *Service) Pull(workspaceID resource.ID) (*task.Task, error) {
	ws, err := s.workspaces.Get(workspaceID)
	if err != nil {
		return nil, err
	}
	return s.tasks.Create(task.CreateOptions{
		Parent:  ws.Resource,
		Command: []string{"state", "pull"},
		Path:    ws.ModulePath(),
		Env:     []string{ws.TerraformEnv()},
	})
}

// ListResources retrieves the state resources for a workspace. Synchronous op.
func (s *Service) ListResources(workspaceID resource.ID) ([]*Resource, error) {
	ws, err := s.workspaces.Get(workspaceID)
	if err != nil {
		return nil, err
	}
	task, err := s.Pull(workspaceID)
	if err != nil {
		return nil, err
	}
	if err := task.Wait(); err != nil {
		return nil, err
	}
	var file StateFile
	if err := json.NewDecoder(task.NewReader()).Decode(&file); err != nil {
		return nil, err
	}
	return file.Resources(ws.Resource), nil
}

// RemoveItems removes items from the state. Aynchronous.
func (s *Service) RemoveItems(workspaceID resource.ID, addrs ...string) (*task.Task, error) {
	// create task invoking "terraform state rm [<addr>...]"
	//
	ws, err := s.workspaces.Get(workspaceID)
	if err != nil {
		return nil, err
	}
	return s.tasks.Create(task.CreateOptions{
		Parent:   ws.Resource,
		Blocking: true,
		Command:  []string{"state", "rm"},
		Args:     addrs,
		Path:     ws.ModulePath(),
		Env:      []string{ws.TerraformEnv()},
	})
}

func (s *Service) Taint(workspaceID resource.ID, addr string) (*task.Task, error) {
	ws, err := s.workspaces.Get(workspaceID)
	if err != nil {
		return nil, err
	}
	return s.tasks.Create(task.CreateOptions{
		Parent:   ws.Resource,
		Blocking: true,
		Command:  []string{"taint"},
		Args:     []string{addr},
		Path:     ws.ModulePath(),
		Env:      []string{ws.TerraformEnv()},
	})
}
