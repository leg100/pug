package run

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type Service struct {
	tasks      *task.Service
	modules    *module.Service
	workspaces *workspace.Service
	// Runs keyed by run ID
	runs map[resource.Resource]*Run
	// Mutex for concurrent read/write of runs
	mu     sync.Mutex
	broker *pubsub.Broker[*Run]
}

type ServiceOptions struct {
	TaskService      *task.Service
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
}

func NewService(opts ServiceOptions) *Service {
	return &Service{
		tasks:      opts.TaskService,
		modules:    opts.ModuleService,
		workspaces: opts.WorkspaceService,
		broker:     pubsub.NewBroker[*Run](),
		runs:       make(map[resource.Resource]*Run),
	}
}

// Create a run, triggering a plan task.
func (s *Service) Create(workspaceID resource.ID, opts CreateOptions) (*Run, *task.Task, error) {
	ws, err := s.workspaces.Get(workspaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("creating run: %w", err)
	}
	mod, err := s.modules.Get(ws.Module().ID)
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
		Path:    mod.Path,
		Command: []string{"plan"},
		Args:    []string{"-input", "false", "-plan", PlanPath(mod, ws, run)},
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

	s.mu.Lock()
	s.runs[run.Resource] = run
	s.mu.Unlock()

	s.broker.Publish(resource.CreatedEvent, run)
	return run, task, nil
}

func (s *Service) afterPlan(mod *module.Module, ws *workspace.Workspace, run *Run) func(*task.Task) {
	return func(plan *task.Task) {
		// Convert binary plan file to json plan file.
		_, err := s.tasks.Create(task.CreateOptions{
			Parent:  run.Resource,
			Path:    mod.Path,
			Command: []string{"show"},
			Args:    []string{"-json", PlanPath(mod, ws, run)},
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
					if _, _, err := s.Apply(run.Resource); err != nil {
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
func (s *Service) Apply(id resource.Resource) (*Run, *task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	run, ok := s.runs[id]
	if !ok {
		return nil, nil, resource.ErrNotFound
	}
	ws, err := s.workspaces.Get(run.Workspace().ID)
	if err != nil {
		return nil, nil, fmt.Errorf("creating run: %w", err)
	}
	mod, err := s.modules.Get(ws.Module().ID)
	if err != nil {
		return nil, nil, fmt.Errorf("creating run: %w", err)
	}

	if run.Status != Planned {
		return nil, nil, fmt.Errorf("run is not in the planned state: %s", run.Status)
	}
	task, err := s.tasks.Create(task.CreateOptions{
		Parent:  run.Resource,
		Path:    mod.Path,
		Command: []string{"apply"},
		Args:    []string{"-input", "false", PlanPath(mod, ws, run)},
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
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("applying run: %w", err)
	}
	run.ApplyTask = &task.Resource
	return run, task, nil
}

func (s *Service) Get(id resource.Resource) (*Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	run, ok := s.runs[id]
	if !ok {
		return nil, resource.ErrNotFound
	}
	return run, nil
}

type ListOptions struct {
	ParentID *resource.ID
}

func (s *Service) List(opts ListOptions) []*Run {
	s.mu.Lock()
	defer s.mu.Unlock()

	var runs []*Run
	for _, run := range s.runs {
		if opts.ParentID != nil {
			if !run.HasAncestor(*opts.ParentID) {
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

func (s *Service) Delete(id resource.Resource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	run, ok := s.runs[id]
	if !ok {
		return resource.ErrNotFound
	}

	if !run.IsFinished() {
		return fmt.Errorf("cannot delete incomplete run")
	}

	delete(s.runs, id)
	s.broker.Publish(resource.DeletedEvent, run)
	return nil
}
