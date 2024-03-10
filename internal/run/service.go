package run

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type Service struct {
	table  *resource.Table[*Run]
	broker *pubsub.Broker[*Run]

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
		broker:     broker,
		tasks:      opts.TaskService,
		modules:    opts.ModuleService,
		workspaces: opts.WorkspaceService,
	}
}

// Create a run, triggering a plan task.
func (s *Service) Create(workspaceID resource.ID, opts CreateOptions) (*Run, *task.Task, error) {
	ws, err := s.workspaces.Get(workspaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("creating run: %w", err)
	}
	mod, err := s.modules.Get(ws.Module().ID())
	if err != nil {
		return nil, nil, fmt.Errorf("creating run: %w", err)
	}
	// Publish an event upon every run status update
	opts.afterUpdate = func(run *Run) {
		s.broker.Publish(resource.UpdatedEvent, run)
	}
	run, err := newRun(mod, ws, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("creating run: %w", err)
	}
	task, err := s.tasks.Create(task.CreateOptions{
		Parent:  run.Resource,
		Path:    mod.Path(),
		Command: []string{"plan"},
		Args:    []string{"-input=false", "-out", run.PlanPath()},
		Env:     []string{ws.TerraformEnv()},
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
		AfterExited: s.afterPlan(mod, ws, run),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("creating plan task: %w", err)
	}
	run.PlanTask = &task.Resource

	s.table.Add(run.ID(), run)
	return run, task, nil
}

func (s *Service) afterPlan(mod *module.Module, ws *workspace.Workspace, run *Run) func(*task.Task) {
	return func(plan *task.Task) {
		// Convert binary plan file to json plan file.
		_, err := s.tasks.Create(task.CreateOptions{
			Parent:  run.Resource,
			Path:    mod.Path(),
			Command: []string{"show"},
			Args:    []string{"-json", run.PlanPath()},
			Env:     []string{ws.TerraformEnv()},
			AfterError: func(t *task.Task) {
				run.setErrored(t.Err)
			},
			AfterCanceled: func(*task.Task) {
				run.updateStatus(Canceled)
			},
			AfterExited: func(t *task.Task) {
				var pfile planFile
				if err := json.NewDecoder(t.NewReader()).Decode(&pfile); err != nil {
					run.setErrored(err)
					return
				}
				if run.addPlan(pfile) {
					if _, _, err := s.Apply(run.ID()); err != nil {
						run.setErrored(err)
						return
					}
				}
			},
		})
		if err != nil {
			run.setErrored(err)
			return
		}
	}
}

// Apply triggers an apply task for a run. The run must be in the planned state.
func (s *Service) Apply(runID resource.ID) (*Run, *task.Task, error) {
	run, err := s.table.Get(runID)
	if err != nil {
		return nil, nil, fmt.Errorf("applying run: %w", err)
	}
	ws, err := s.workspaces.Get(run.Workspace().ID())
	if err != nil {
		return nil, nil, fmt.Errorf("applying run: %w", err)
	}

	if run.Status != Planned {
		return nil, nil, fmt.Errorf("run is not in the planned state: %s", run.Status)
	}
	task, err := s.tasks.Create(task.CreateOptions{
		Parent:  run.Resource,
		Path:    run.ModulePath(),
		Command: []string{"apply"},
		Args:    []string{"-input=false", run.PlanPath()},
		Env:     []string{ws.TerraformEnv()},
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
				// log error
				return
			}
			report, err := parseApplyOutput(string(out))
			if err != nil {
				// log error
				return
			}
			run.ApplyReport = report
			run.updateStatus(Applied)
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("applying run: %w", err)
	}
	run.ApplyTask = &task.Resource
	return run, task, nil
}

func (s *Service) Get(runID resource.ID) (*Run, error) {
	return s.table.Get(runID)
}

type ListOptions struct {
	ParentID resource.ID
}

func (s *Service) List(opts ListOptions) []*Run {
	var runs []*Run
	for _, run := range s.table.List() {
		if opts.ParentID != resource.NilID {
			if !run.HasAncestor(opts.ParentID) {
				continue
			}
		}
		runs = append(runs, run)
	}
	return runs
}

func (s *Service) Subscribe(ctx context.Context) (<-chan resource.Event[*Run], func()) {
	return s.broker.Subscribe(ctx)
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
