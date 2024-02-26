package run

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type Service struct {
	tasks      *task.Service
	workspaces *workspace.Service
	modules    *module.Service
	// Runs keyed by run ID
	runs map[resource.Resource]*Run
	// Mutex for concurrent read/write of runs
	mu     sync.Mutex
	broker *pubsub.Broker[*Run]
}

// Create a run, triggering a plan task.
func (s *Service) Create(id workspace.ID, opts CreateOptions) (*Run, *task.Task, error) {
	ws, err := s.workspaces.Get(id)
	if err != nil {
		return nil, nil, fmt.Errorf("creating run: %w", err)
	}
	run, err := newRun(ws, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("creating run: %w", err)
	}
	task, err := s.workspaces.CreateTask(ws.Resource, task.CreateOptions{
		Command:     []string{"plan"},
		Args:        []string{"-input", "false", "-plan", run.PlanPath()},
		AfterExited: s.afterPlan(run),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("creating plan task: %w", err)
	}

	s.mu.Lock()
	s.runs[run.Resource] = run
	s.mu.Unlock()

	s.broker.Publish(resource.CreatedEvent, run)
	return run, task, nil
}

func (s *Service) afterPlan(run *Run) func(*task.Task) {
	return func(plan *task.Task) {
		// Convert binary plan file to json plan file.
		convert, err := s.workspaces.CreateTask(run.Workspace, task.CreateOptions{
			Command: []string{"show"},
			Args:    []string{"-json", run.PlanPath()},
		})
		if err != nil {
			run.setErrored(err)
			return
		}
		// TODO: make task above synchronous
		var pfile planFile
		if err = json.NewDecoder(convert.NewReader()).Decode(&pfile); err != nil {
			run.setErrored(err)
			return
		}
		run.PlanReport = pfile.resourceChanges()
		if !run.PlanReport.HasChanges() {
			run.updateStatus(PlannedAndFinished)
			return
		}
		run.updateStatus(Planned)
		if run.AutoApply {
			_, _, err = s.Apply(run.Resource)
			if err != nil {
				run.setErrored(err)
				return
			}
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
	if run.Status != Planned {
		return nil, nil, fmt.Errorf("run is not in the planned state: %s", run.Status)
	}
	task, err := s.workspaces.CreateTask(run.Parent, task.CreateOptions{
		Command: []string{"apply"},
		Args:    []string{"-input", "false", run.PlanPath()},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("applying run: %w", err)
	}
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

func (s *Service) List(id workspace.ID) []*Run {
	s.mu.Lock()
	defer s.mu.Unlock()

	var runs []*Run
	for _, run := range s.runs {
		if run.Workspace == id {
			runs = append(runs, run)
		}
	}
	return runs
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

const MetadataKey = "run"

// Create task for a run.
func (s *Service) CreateTask(run *Run, opts task.CreateOptions) (*task.Task, error) {
	return s.workspaces.CreateTask(run.Parent, opts)
}
