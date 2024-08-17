package workspace

import (
	"fmt"

	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

type Service struct {
	table  workspaceTable
	logger logging.Interface

	modules modules
	tasks   *task.Service

	*pubsub.Broker[*Workspace]
	*reloader
}

type ServiceOptions struct {
	Tasks   *task.Service
	Modules *module.Service
	Logger  logging.Interface
}

type workspaceTable interface {
	Add(id resource.ID, row *Workspace)
	Update(id resource.ID, updater func(existing *Workspace) error) (*Workspace, error)
	Delete(id resource.ID)
	Get(id resource.ID) (*Workspace, error)
	List() []*Workspace
}

type modules interface {
	Get(id resource.ID) (*module.Module, error)
	GetByPath(path string) (*module.Module, error)
	SetCurrent(moduleID, workspaceID resource.ID) error
	Reload() ([]string, []string, error)
	List() []*module.Module
}

func NewService(opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*Workspace](opts.Logger)
	table := resource.NewTable(broker)

	opts.Logger.AddEnricher(&logEnricher{table: table})

	s := &Service{
		Broker:  broker,
		table:   table,
		modules: opts.Modules,
		tasks:   opts.Tasks,
		logger:  opts.Logger,
	}
	s.reloader = &reloader{s}
	return s
}

// LoadWorkspacesUponModuleLoad automatically loads workspaces for a module
// whenever:
// * a new module is loaded into pug for the first time
// * an existing module is updated and does not yet have a current workspace.
//
// TODO: "load" is ambiguous, it often means the opposite of save, i.e. read
// from a system, whereas what is intended is to save or add workspaces to pug.
func (s *Service) LoadWorkspacesUponModuleLoad(sub <-chan resource.Event[*module.Module]) {
	reload := func(moduleID resource.ID) error {
		spec, err := s.Reload(moduleID)
		if err != nil {
			return err
		}
		_, err = s.tasks.Create(spec)
		return err
	}

	for event := range sub {
		switch event.Type {
		case resource.CreatedEvent:
			if err := reload(event.Payload.ID); err != nil {
				s.logger.Error("reloading workspaces", "module", event.Payload)
			}
		case resource.UpdatedEvent:
			if event.Payload.CurrentWorkspaceID != nil {
				// Module already has a current workspace; no need to reload
				// workspaces
				continue
			}
			if err := reload(event.Payload.ID); err != nil {
				s.logger.Error("reloading workspaces", "module", event.Payload)
			}
		}
	}
}

// Create a workspace. Asynchronous.
func (s *Service) Create(path, name string) (task.Spec, error) {
	mod, err := s.modules.GetByPath(path)
	if err != nil {
		return task.Spec{}, err
	}
	ws, err := New(mod, name)
	if err != nil {
		return task.Spec{}, err
	}
	return task.Spec{
		Parent:  mod,
		Path:    mod.Path,
		Command: []string{"workspace", "new"},
		Args:    []string{name},
		AfterExited: func(*task.Task) {
			s.table.Add(ws.ID, ws)
			// `workspace new` implicitly makes the created workspace the
			// *current* workspace, so better tell pug that too.
			if err := s.modules.SetCurrent(mod.ID, ws.ID); err != nil {
				s.logger.Error("creating workspace: %w", err)
			}
		},
	}, nil
}

func (s *Service) Get(workspaceID resource.ID) (*Workspace, error) {
	return s.table.Get(workspaceID)
}

func (s *Service) GetByName(moduleID resource.ID, name string) (*Workspace, error) {
	for _, ws := range s.table.List() {
		if ws.ModuleID() == moduleID && ws.Name == name {
			return ws, nil
		}
	}
	return nil, resource.ErrNotFound
}

type ListOptions struct {
	// Filter by ID of workspace's module. If zero value then no filtering is
	// performed.
	ModuleID resource.ID
}

func (s *Service) List(opts ListOptions) []*Workspace {
	var existing []*Workspace
	for _, ws := range s.table.List() {
		// If opts.ModuleID is zero value then HasAncestor returns true.
		if !ws.HasAncestor(opts.ModuleID) {
			continue
		}
		existing = append(existing, ws)
	}
	return existing
}

// SelectWorkspace runs the `terraform workspace select <workspace_name>`
// command, which sets the current workspace for the module. Once that's
// finished it then updates the current workspace in pug itself too.
func (s *Service) SelectWorkspace(moduleID, workspaceID resource.ID) error {
	ws, err := s.table.Get(workspaceID)
	if err != nil {
		return err
	}
	mod, err := s.modules.Get(ws.ModuleID())
	if err != nil {
		return err
	}
	// Create task to immediately set workspace as current workspace for module.
	_, err = s.tasks.Create(task.Spec{
		Parent:    mod,
		Path:      mod.Path,
		Command:   []string{"workspace", "select"},
		Args:      []string{ws.Name},
		Immediate: true,
		Wait:      true,
		BeforeExited: func(t *task.Task) (task.Summary, error) {
			// Now the terraform command has finished, update the current
			// workspace in pug as well.
			err := s.modules.SetCurrent(moduleID, workspaceID)
			return nil, err
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// Delete a workspace. Asynchronous.
func (s *Service) Delete(workspaceID resource.ID) (task.Spec, error) {
	ws, err := s.table.Get(workspaceID)
	if err != nil {
		return task.Spec{}, fmt.Errorf("deleting workspace: %w", err)
	}
	mod, err := s.modules.Get(ws.ModuleID())
	if err != nil {
		return task.Spec{}, err
	}
	return task.Spec{
		Parent:   mod,
		Path:     mod.Path,
		Command:  []string{"workspace", "delete"},
		Args:     []string{ws.Name},
		Blocking: true,
		AfterExited: func(*task.Task) {
			s.table.Delete(ws.ID)
		},
	}, nil
}

// Cost creates a task that retrieves a breakdown of the costs of the
// infrastructure deployed by the workspace.
// func (s *Service) Cost(workspaceID resource.ID) (*task.Task, error) {
// 	ws, err := s.table.Get(workspaceID)
// 	if err != nil {
// 		return nil, fmt.Errorf("costing workspace: %w", err)
// 	}
// 	return s.createTask(ws, task.CreateOptions{
// 		Command: []string{"infracost"},
// 		Args:    []string{"breakdown", "-p", ws.ModulePath(), "--terraform-workspace", ws.Name},
// 		AfterExited: func(t *task.Task) {
// 			cost, err := parseInfracostOutput(t.NewReader())
// 			if err != nil {
// 				s.logger.Error("parsing infracost output", "error", err, "workspace", ws)
// 			}
// 			t.SetSummary(cost)
// 		},
// 	})
// }
