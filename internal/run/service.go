package run

import (
	"fmt"
	"slices"

	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type Service struct {
	table  *resource.Table[*Run]
	logger logging.Interface

	tasks      *task.Service
	modules    moduleGetter
	workspaces workspaceGetter
	states     *state.Service

	disableReloadAfterApply bool

	*factory
	*pubsub.Broker[*Run]
}

type ServiceOptions struct {
	TaskService             *task.Service
	ModuleService           *module.Service
	WorkspaceService        *workspace.Service
	StateService            *state.Service
	DisableReloadAfterApply bool
	DataDir                 string
	Logger                  logging.Interface
	Terragrunt              bool
}

type moduleGetter interface {
	Get(moduleID resource.ID) (*module.Module, error)
}

type workspaceGetter interface {
	Get(workspaceID resource.ID) (*workspace.Workspace, error)
}

func NewService(opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*Run](opts.Logger)
	return &Service{
		table:                   resource.NewTable(broker),
		Broker:                  broker,
		tasks:                   opts.TaskService,
		modules:                 opts.ModuleService,
		workspaces:              opts.WorkspaceService,
		states:                  opts.StateService,
		disableReloadAfterApply: opts.DisableReloadAfterApply,
		logger:                  opts.Logger,
		factory: &factory{
			dataDir:    opts.DataDir,
			modules:    opts.ModuleService,
			workspaces: opts.WorkspaceService,
			broker:     broker,
			terragrunt: opts.Terragrunt,
		},
	}
}

// Plan creates a plan task.
func (s *Service) Plan(workspaceID resource.ID, opts CreateOptions) (*task.Task, error) {
	task, err := s.plan(workspaceID, opts)
	if err != nil {
		s.logger.Error("creating plan task", "error", err)
		return nil, err
	}
	return task, nil
}

func (s *Service) plan(workspaceID resource.ID, opts CreateOptions) (*task.Task, error) {
	run, err := s.newRun(workspaceID, opts)
	if err != nil {
		return nil, err
	}
	task, err := s.createTask(run, task.CreateOptions{
		Command:  []string{"plan"},
		Args:     run.planArgs(),
		Blocking: true,
		AfterQueued: func(*task.Task) {
			run.updateStatus(PlanQueued)
		},
		AfterRunning: func(*task.Task) {
			run.updateStatus(Planning)
		},
		AfterError: func(t *task.Task) {
			run.updateStatus(Errored)
		},
		AfterCanceled: func(*task.Task) {
			run.updateStatus(Canceled)
		},
		AfterExited: func(t *task.Task) {
			if err := run.finishPlan(t.NewReader()); err != nil {
				s.logger.Error("finishing plan", "error", err, "run", run)
			}
		},
	})
	if err != nil {
		return nil, err
	}
	s.table.Add(run.ID, run)
	return task, nil
}

// Apply either auto-applies a new run, or it applies the plan from an existing
// run. To auto-apply a new run, specify the ID of the workspace on which to
// create the new run and provide opts. If applying the plan from an existing
// run, specify the ID of the run, and ensure opts is nil.
func (s *Service) Apply(id resource.ID, opts *CreateOptions) (*task.Task, error) {
	var (
		run *Run
		err error
	)
	if opts != nil {
		// Create new run
		opts.applyOnly = true
		run, err = s.newRun(id, *opts)
	} else {
		// Apply plan from existing run.
		run, err = s.table.Get(id)
		if run != nil && run.Status != Planned {
			err = fmt.Errorf("run is not in the planned state: %s", run.Status)
		}
	}
	if err != nil {
		return nil, err
	}

	task, err := s.createTask(run, task.CreateOptions{
		Command:  []string{"apply"},
		Args:     run.applyArgs(),
		Blocking: true,
		DependsOn: 
		AfterQueued: func(*task.Task) {
			run.updateStatus(ApplyQueued)
		},
		AfterRunning: func(*task.Task) {
			run.updateStatus(Applying)
		},
		AfterError: func(*task.Task) {
			run.updateStatus(Errored)
		},
		AfterCanceled: func(*task.Task) {
			run.updateStatus(Canceled)
		},
		AfterExited: func(t *task.Task) {
			if err := run.finishApply(t.NewReader()); err != nil {
				s.logger.Error("finishing apply", "error", err, "run", run)
				return
			}

			if !s.disableReloadAfterApply {
				s.states.Reload(run.WorkspaceID())
			}
		},
	})
	if err != nil {
		s.logger.Error("creating an apply task", "error", err)
		return nil, err
	}
	s.table.Add(run.ID, run)
	return task, nil
}

// MultiApply creates multiple apply tasks. This operation is carried out in a
// special routine because applies may depend upon other applies, in which case
// the IDs must be ordered accordingly.
func (s *Service) MultiApply(opts *CreateOptions, ids ...resource.ID) (*task.Group, error) {
	for _, id := range ids {
		switch id.Kind {
		case resource.Workspace:
			ws, err := h.WorkspaceService.Get(id)
			if err != nil {
				// report error
			}
			ws.Module()
		case resource.Run:
		}

	}
			group, err := h.TaskService.CreateGroup("apply", fn, ids...)
	var (
		run *Run
		err error
	)
	if opts != nil {
		// Create new run
		opts.applyOnly = true
		run, err = s.newRun(id, *opts)
	} else {
		// Apply plan from existing run.
		run, err = s.table.Get(id)
		if run != nil && run.Status != Planned {
			err = fmt.Errorf("run is not in the planned state: %s", run.Status)
		}
	}
	if err != nil {
		return nil, err
	}

	task, err := s.createTask(run, task.CreateOptions{
		Command:  []string{"apply"},
		Args:     run.applyArgs(),
		Blocking: true,
		AfterQueued: func(*task.Task) {
			run.updateStatus(ApplyQueued)
		},
		AfterRunning: func(*task.Task) {
			run.updateStatus(Applying)
		},
		AfterError: func(*task.Task) {
			run.updateStatus(Errored)
		},
		AfterCanceled: func(*task.Task) {
			run.updateStatus(Canceled)
		},
		AfterExited: func(t *task.Task) {
			if err := run.finishApply(t.NewReader()); err != nil {
				s.logger.Error("finishing apply", "error", err, "run", run)
				return
			}

			if !s.disableReloadAfterApply {
				s.states.Reload(run.WorkspaceID())
			}
		},
	})
	if err != nil {
		s.logger.Error("creating an apply task", "error", err)
		return nil, err
	}
	s.table.Add(run.ID, run)
	return task, nil
}

func (s *Service) Get(runID resource.ID) (*Run, error) {
	return s.table.Get(runID)
}

type ListOptions struct {
	// Filter runs by those that belong to the given ancestor, e.g. a module or workspace
	AncestorID resource.ID
	// Filter runs by status: match run if it has one of these statuses.
	Status []Status
	// Order runs by oldest first (true), or newest first (false)
	Oldest bool
}

func (s *Service) List(opts ListOptions) []*Run {
	runs := s.table.List()

	// Filter list according to options
	var i int
	for _, r := range runs {
		if opts.Status != nil {
			if !slices.Contains(opts.Status, r.Status) {
				continue
			}
		}
		if opts.AncestorID != resource.GlobalID {
			if !r.HasAncestor(opts.AncestorID) {
				continue
			}
		}
		runs[i] = r
		i++
	}
	runs = runs[:i]

	// Sort list according to options
	slices.SortFunc(runs, func(a, b *Run) int {
		cmp := a.Updated.Compare(b.Updated)
		if opts.Oldest {
			return cmp
		}
		return -cmp
	})

	return runs
}

func (s *Service) Delete(id resource.ID) error {
	if err := s.delete(id); err != nil {
		return err
	}
	s.logger.Info("deleted run", "id", id)
	return nil
}

func (s *Service) delete(id resource.ID) error {
	run, err := s.table.Get(id)
	if err != nil {
		return fmt.Errorf("deleting run: %w", err)
	}

	if !run.IsFinished() {
		return fmt.Errorf("cannot delete incomplete run")
	}
	s.table.Delete(id)
	return nil
}

// TODO: move this logic into task.Create
func (s *Service) createTask(run *Run, opts task.CreateOptions) (*task.Task, error) {
	opts.Parent = run

	ws, err := s.workspaces.Get(run.WorkspaceID())
	if err != nil {
		return nil, err
	}
	opts.Env = []string{ws.TerraformEnv()}

	mod, err := s.modules.Get(ws.ModuleID())
	if err != nil {
		return nil, err
	}
	opts.Path = mod.Path

	return s.tasks.Create(opts)
}
