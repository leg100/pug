package run

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type Service struct {
	Broker *pubsub.Broker[*Run]
	table  *resource.Table[*Run]

	tasks      *task.Service
	modules    *module.Service
	workspaces *workspace.Service
}

type ServiceOptions struct {
	TaskService      *task.Service
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
}

func NewService(opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*Run]()
	return &Service{
		table:      resource.NewTable[*Run](broker),
		Broker:     broker,
		tasks:      opts.TaskService,
		modules:    opts.ModuleService,
		workspaces: opts.WorkspaceService,
	}
}

// Create a run.
func (s *Service) Create(workspaceID resource.ID, opts CreateOptions) (*Run, error) {
	ws, err := s.workspaces.Get(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("creating run: %w", err)
	}
	mod, err := s.modules.Get(ws.Module().ID())
	if err != nil {
		return nil, fmt.Errorf("creating run: %w", err)
	}
	// Publish an event upon every run status update
	opts.afterUpdate = func(run *Run) {
		s.Broker.Publish(resource.UpdatedEvent, run)
	}
	run, err := newRun(mod, ws, opts)
	if err != nil {
		return nil, fmt.Errorf("creating run: %w", err)
	}
	s.table.Add(run.ID(), run)
	return run, nil
}

// Trigger a plan task for a run. Only to be called by the scheduler.
func (s *Service) plan(run *Run) (*task.Task, error) {
	task, err := s.createTask(run, task.CreateOptions{
		Command: []string{"plan"},
		Args:    []string{"-lock=false", "-input=false", "-out", run.PlanPath()},
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
				run.updateStatus(PlannedAndFinished)
				return
			}
			run.updateStatus(Planned)
			if run.AutoApply {
				if _, err := s.Apply(run.ID()); err != nil {
					run.setErrored(err)
					return
				}
			}
		},
	})
	if err != nil {
		return nil, fmt.Errorf("creating plan task: %w", err)
	}
	return task, err
}

// Apply triggers an apply task for a run. The run must be in the planned state.
func (s *Service) Apply(runID resource.ID) (*task.Task, error) {
	run, err := s.table.Get(runID)
	if err != nil {
		return nil, fmt.Errorf("applying run: %w", err)
	}
	ws, err := s.workspaces.Get(run.Workspace().ID())
	if err != nil {
		return nil, fmt.Errorf("applying run: %w", err)
	}

	if run.Status != Planned {
		return nil, fmt.Errorf("run is not in the planned state: %s", run.Status)
	}
	task, err := s.tasks.Create(task.CreateOptions{
		Parent:   run.Resource,
		Path:     run.ModulePath(),
		Blocking: true,
		Command:  []string{"apply"},
		Args:     []string{"-input=false", run.PlanPath()},
		Env:      []string{ws.TerraformEnv()},
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
		},
	})
	if err != nil {
		return nil, fmt.Errorf("applying run: %w", err)
	}
	run.ApplyTask = &task.Resource
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
	// Filter runs by plan-only (true), not plan-only (false), or either (nil)
	PlanOnly *bool
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
		if opts.PlanOnly != nil {
			if *opts.PlanOnly {
				if !r.PlanOnly {
					// Exclude runs that are not plan-only
					continue
				}
			} else if r.PlanOnly {
				// Exclude plan-only runs
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
	opts.Path = run.ModulePath()
	opts.Env = []string{workspace.TerraformEnv(run.WorkspaceName())}
	return s.tasks.Create(opts)
}
