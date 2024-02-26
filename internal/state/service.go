package state

import (
	"encoding/json"
	"fmt"

	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type Service struct {
	tasks      *task.Service
	workspaces *workspace.Service
}

// Get asynchronously retrieves the state for a workspace.
func (s *Service) Get(id workspace.ID) (*task.Task, error) {
	ws, err := s.workspaces.Get(id)
	if err != nil {
		return nil, fmt.Errorf("pulling state: %w", err)
	}
	return s.tasks.Create(task.CreateOptions{
		Kind: workspace.ShowStateTask,
		Args: []string{"state", "pull"},
		Path: ws.ModulePath,
	})
}

// ListResources retrieves the resources for a workspace.
func (s *Service) ListResources(id workspace.ID) ([]Resource, error) {
	get, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("listing resources: %w", err)
	}
	// TODO: make the above a synchronous op
	var file File
	if err := json.NewDecoder(get.NewReader()).Decode(&file); err != nil {
		return nil, err
	}
	return file.Resources, nil
}

// RemoveItems removes items from the state. Aynchronous.
func (s *Service) RemoveItems(id workspace.ID, addrs ...string) (*task.Task, error) {
	// create task invoking "terraform state rm [<addr>...]"
	//
	ws, err := s.workspaces.Get(id)
	if err != nil {
		return nil, fmt.Errorf("pulling state: %w", err)
	}
	return s.tasks.Create(task.CreateOptions{
		Kind: workspace.RemoveStateItemsTask,
		Args: append([]string{"state", "rm"}, addrs...),
		Path: ws.ModulePath,
	})
}

func (s *Service) Taint(id workspace.ID, addr string) (*task.Task, error) {
	ws, err := s.workspaces.Get(id)
	if err != nil {
		return nil, fmt.Errorf("pulling state: %w", err)
	}
	return s.tasks.Create(task.CreateOptions{
		Args: []string{"taint", addr},
		Path: ws.ModulePath,
		// Workspace: id
	})
}
