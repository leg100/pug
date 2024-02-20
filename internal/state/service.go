package state

import (
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type Service struct {
	tasks      *task.Service
	workspaces *workspace.Service
}

// Get asynchronously retrieves the state for a workspace.
func (s *Service) Get(id workspace.ID) (*task.Task, error) {
}

// RemoveItems removes items from the state. This is an asynchronous operation.
func (s *Service) RemoveItems(id workspace.ID, addresses ...string) (*task.Task, error) {
	// create task invoking "terraform state rm [<addr>...]"
	//
}
