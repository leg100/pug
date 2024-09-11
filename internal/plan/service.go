package plan

import (
	"fmt"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type Service struct {
	table  *resource.Table[*plan]
	logger logging.Interface

	tasks      *task.Service
	modules    moduleGetter
	workspaces workspaceGetter
	states     *state.Service

	*factory
	*pubsub.Broker[*plan]
}

type ServiceOptions struct {
	Tasks      *task.Service
	Modules    *module.Service
	Workspaces *workspace.Service
	States     *state.Service
	DataDir    string
	Workdir    internal.Workdir
	Logger     logging.Interface
	Terragrunt bool
}

type moduleGetter interface {
	Get(moduleID resource.ID) (*module.Module, error)
}

type workspaceGetter interface {
	Get(workspaceID resource.ID) (*workspace.Workspace, error)
}

func NewService(opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*plan](opts.Logger)
	return &Service{
		table:      resource.NewTable(broker),
		Broker:     broker,
		tasks:      opts.Tasks,
		modules:    opts.Modules,
		workspaces: opts.Workspaces,
		states:     opts.States,
		logger:     opts.Logger,
		factory: &factory{
			dataDir:    opts.DataDir,
			workdir:    opts.Workdir,
			modules:    opts.Modules,
			workspaces: opts.Workspaces,
			broker:     broker,
			terragrunt: opts.Terragrunt,
		},
	}
}

// ReloadAfterApply creates a state reload task whenever an apply task
// successfully finishes.
func (s *Service) ReloadAfterApply(sub <-chan resource.Event[*task.Task]) {
	for event := range sub {
		switch event.Type {
		case resource.UpdatedEvent:
			if event.Payload.State != task.Exited {
				continue
			}
			if event.Payload.Identifier != ApplyTask {
				continue
			}
			workspaceID := event.Payload.WorkspaceID
			if workspaceID == nil {
				continue
			}
			if _, err := s.states.CreateReloadTask(*workspaceID); err != nil {
				s.logger.Error("reloading state after apply", "error", err, "workspace", *workspaceID)
				continue
			}
			s.logger.Debug("reloading state after apply", "workspace", *workspaceID)
		}
	}
}

// Plan creates a task spec to create a plan, i.e. `terraform plan -out
// plan.file`.
func (s *Service) Plan(workspaceID resource.ID, opts CreateOptions) (task.Spec, error) {
	opts.planFile = true
	plan, err := s.newPlan(workspaceID, opts)
	if err != nil {
		s.logger.Error("creating plan spec", "error", err)
		return task.Spec{}, err
	}
	s.table.Add(plan.ID, plan)

	return plan.planTaskSpec(), nil
}

// Apply creates a task spec to auto-apply a plan, i.e. `terraform apply`. To
// apply an existing plan, see ApplyPlan.
func (s *Service) Apply(workspaceID resource.ID, opts CreateOptions) (task.Spec, error) {
	plan, err := s.newPlan(workspaceID, opts)
	if err != nil {
		return task.Spec{}, err
	}
	return plan.applyTaskSpec()
}

// ApplyPlan creates a task spec to apply an existing plan, i.e. `terraform
// apply existing.plan`. The taskID is the ID of a plan task, which must have
// finished successfully.
func (s *Service) ApplyPlan(taskID resource.ID) (task.Spec, error) {
	planTask, err := s.tasks.Get(taskID)
	if err != nil {
		return task.Spec{}, err
	}
	if planTask.State != task.Exited {
		return task.Spec{}, fmt.Errorf("plan task is not in the exited state: %s", planTask.State)
	}
	plan, err := s.getByTaskID(taskID)
	if err != nil {
		return task.Spec{}, err
	}
	return plan.applyTaskSpec()
}

func (s *Service) Get(runID resource.ID) (*plan, error) {
	return s.table.Get(runID)
}

func (s *Service) getByTaskID(taskID resource.ID) (*plan, error) {
	for _, plan := range s.List() {
		if plan.taskID != nil && *plan.taskID == taskID {
			return plan, nil
		}
	}
	return nil, fmt.Errorf("task is not associated with a plan: %w", resource.ErrNotFound)
}

func (s *Service) List() []*plan {
	return s.table.List()
}
