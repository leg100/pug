package state

import (
	"errors"

	"github.com/leg100/pug/internal/resource"
)

type State struct {
	WorkspaceID resource.ID
	Resources   map[ResourceAddress]*Resource
	State       StateStatus
}

type StateStatus string

const (
	// Idle indicates the resource is idle (no tasks are currently operating on
	// it).
	IdleState StateStatus = "idle"
	// Removing indicates the resource is in the process of being removed.
	ReloadingState = "reloading"
)

func EmptyState(workspaceID resource.ID) *State {
	return &State{
		WorkspaceID: workspaceID,
		State:       ReloadingState,
	}
}

func NewState(workspaceID resource.ID, file StateFile) *State {
	return &State{
		WorkspaceID: workspaceID,
		Resources:   getResourcesFromFile(file),
		State:       IdleState,
	}
}

func (s *State) startReload() error {
	if s.State == ReloadingState {
		return errors.New("state is already reloading")
	}
	s.State = ReloadingState
	return nil
}
