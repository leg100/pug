package run

import (
	"context"
	"fmt"
	"io"
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
	Broker *pubsub.Broker[*Run]

	table  *resource.Table[*Run]
	logger logging.Interface

	tasks      *task.Service
	modules    *module.Service
	workspaces *workspace.Service
	states     *state.Service

	disableReloadAfterApply bool
}

type ServiceOptions struct {
	TaskService             *task.Service
	ModuleService           *module.Service
	WorkspaceService        *workspace.Service
	StateService            *state.Service
	DisableReloadAfterApply bool
	Logger                  logging.Interface
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
	}
}

// Create a run.
func (s *Service) Create(workspaceID resource.ID, opts CreateOptions) (*Run, error) {
	run, err := s.create(workspaceID, opts)
	if err != nil {
		s.logger.Error("creating run", "error", err, "workspace_id", workspaceID)
		return nil, err
	}
	s.logger.Info("created run", "run", run)
	return run, nil
}

func (s *Service) create(workspaceID resource.ID, opts CreateOptions) (*Run, error) {
	ws, err := s.workspaces.Get(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("retrieving workspace: %w", err)
	}
	mod, err := s.modules.Get(ws.ModuleID())
	if err != nil {
		return nil, fmt.Errorf("workspace module: %w", err)
	}
	// Publish an event upon every run status update
	opts.afterUpdate = func(run *Run) {
		s.Broker.Publish(resource.UpdatedEvent, run)
	}
	run, err := newRun(mod, ws, opts)
	if err != nil {
		return nil, fmt.Errorf("constructing run: %w", err)
	}
	s.table.Add(run.ID, run)
	return run, nil
}

// Create a plan task for a run. Only to be called by the scheduler.
func (s *Service) plan(run *Run) (*task.Task, error) {
	task, err := s.createTask(run, task.CreateOptions{
		Command:  []string{"plan"},
		Args:     run.PlanArgs(),
		Blocking: true,
		AfterQueued: func(*task.Task) {
			run.updateStatus(PlanQueued)
		},
		AfterRunning: func(*task.Task) {
			run.updateStatus(Planning)
		},
		AfterError: func(t *task.Task) {
			run.setErrored(t.Err)
		},
		AfterCanceled: func(*task.Task) {
			run.updateStatus(Canceled)
		},
		AfterExited: func(t *task.Task) {
			out, err := io.ReadAll(t.NewReader())
			if err != nil {
				run.setErrored(err)
				return
			}
			changes, report, err := parsePlanReport(string(out))
			if err != nil {
				run.setErrored(err)
				return
			}
			run.PlanReport = report

			// Determine status and whether to automatically proceed to apply
			if !changes {
				run.updateStatus(NoChanges)
				return
			}
			run.updateStatus(Planned)
			if run.AutoApply {
				if _, err := s.Apply(run.ID); err != nil {
					run.setErrored(err)
					return
				}
			}
			s.logger.Info("created plan", "run", run, "changes", run.PlanReport)
		},
		AfterFinish: func(t *task.Task) {
			if t.Err != nil {
				s.logger.Error("creating plan", "error", t.Err, "run", run)
			}
		},
	})
	if err != nil {
		s.logger.Error("creating plan task", "error", err, "run", run)
		return nil, err
	}
	return task, nil
}

// Apply creates an apply task for a run. The run must be in the planned state,
// and it must be the current run for its workspace.
func (s *Service) Apply(runID resource.ID) (*task.Task, error) {
	task, err := s.apply(runID)
	if err != nil {
		s.logger.Error("applying plan", "error", err, "run_id", runID)
		return nil, err
	}
	return task, err
}

func (s *Service) apply(runID resource.ID) (*task.Task, error) {
	run, err := s.table.Get(runID)
	if err != nil {
		return nil, err
	}
	if run.Status != Planned {
		return nil, fmt.Errorf("run is not in the planned state: %s", run.Status)
	}
	ws, err := s.workspaces.Get(run.WorkspaceID())
	if err != nil {
		return nil, err
	}
	if ws.CurrentRunID == nil || *ws.CurrentRunID != runID {
		return nil, fmt.Errorf("run is not the current run for its workspace: current run: %s", ws.CurrentRunID)
	}
	task, err := s.createTask(run, task.CreateOptions{
		Command:  []string{"apply"},
		Args:     []string{"-input=false", run.PlanPath()},
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
			// TODO: mark all workspace runs in the planned state as stale
			//
			out, err := io.ReadAll(t.NewReader())
			if err != nil {
				run.setErrored(err)
				return
			}
			report, err := parseApplyReport(string(out))
			if err != nil {
				run.setErrored(err)
				return
			}
			run.ApplyReport = report
			run.updateStatus(Applied)

			if !s.disableReloadAfterApply {
				s.states.Reload(run.WorkspaceID())
			}
		},
		AfterFinish: func(t *task.Task) {
			if t.Err != nil {
				s.logger.Error("applying plan", "error", t.Err, "run", run)
			} else {
				s.logger.Info("applied plan", "run", run)
			}
		},
	})
	if err != nil {
		return nil, err
	}
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

func (s *Service) Subscribe(ctx context.Context) <-chan resource.Event[*Run] {
	return s.Broker.Subscribe(ctx)
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

func (s *Service) createTask(run *Run, opts task.CreateOptions) (*task.Task, error) {
	opts.Parent = run.Resource

	ws, err := s.workspaces.Get(run.WorkspaceID())
	if err != nil {
		return nil, err
	}
	opts.Env = []string{workspace.TerraformEnv(ws.Name)}

	mod, err := s.modules.Get(ws.ModuleID())
	if err != nil {
		return nil, err
	}
	opts.Path = mod.Path

	return s.tasks.Create(opts)
}
