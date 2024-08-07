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
	Tasks                   *task.Service
	Modules                 *module.Service
	Workspaces              *workspace.Service
	States                  *state.Service
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
		tasks:                   opts.Tasks,
		modules:                 opts.Modules,
		workspaces:              opts.Workspaces,
		states:                  opts.States,
		disableReloadAfterApply: opts.DisableReloadAfterApply,
		logger:                  opts.Logger,
		factory: &factory{
			dataDir:    opts.DataDir,
			modules:    opts.Modules,
			workspaces: opts.Workspaces,
			broker:     broker,
			terragrunt: opts.Terragrunt,
		},
	}
}

// Plan creates a plan task.
func (s *Service) Plan(workspaceID resource.ID, opts CreateOptions) (task.Spec, error) {
	spec, err := s.plan(workspaceID, opts)
	if err != nil {
		s.logger.Error("creating plan spec", "error", err)
		return task.Spec{}, err
	}
	return spec, nil
}

func (s *Service) plan(workspaceID resource.ID, opts CreateOptions) (task.Spec, error) {
	run, err := s.newRun(workspaceID, opts)
	if err != nil {
		return task.Spec{}, err
	}
	s.table.Add(run.ID, run)

	return task.Spec{
		Parent:  run,
		Path:    run.ModulePath(),
		Env:     []string{workspace.TerraformEnv(run.WorkspaceName())},
		Command: []string{"plan"},
		Args:    run.planArgs(),
		// TODO: explain why plan is blocking (?)
		Blocking:    true,
		Description: PlanTaskDescription(opts.Destroy),
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
			if err := run.finishPlan(t.NewReader(false)); err != nil {
				s.logger.Error("finishing plan", "error", err, "run", run)
			}
		},
	}, nil
}

// Apply creates a task for a terraform apply.
//
// If opts is non-nil, then a new run is created and auto-applied without creating a
// plan file. The ID must be the workspace ID on which to create the run.
//
// If opts is nil, then it will apply an existing plan. The ID must specify an
// existing run that has successfully created a plan.
func (s *Service) Apply(id resource.ID, opts *CreateOptions) (task.Spec, error) {
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
		return task.Spec{}, err
	}
	s.table.Add(run.ID, run)

	spec := task.Spec{
		Parent:      run,
		Path:        run.ModulePath(),
		Command:     []string{"apply"},
		Args:        run.applyArgs(),
		Env:         []string{workspace.TerraformEnv(run.WorkspaceName())},
		Blocking:    true,
		Description: ApplyTaskDescription(run.Destroy),
		// If terragrunt is in use then respect module dependencies.
		RespectModuleDependencies: s.terragrunt,
		// Module dependencies are reversed for a destroy.
		InverseDependencyOrder: run.Destroy,
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
			if err := run.finishApply(t.NewReader(false)); err != nil {
				s.logger.Error("finishing apply", "error", err, "run", run)
				return
			}

			if !s.disableReloadAfterApply {
				spec, err := s.states.Reload(run.WorkspaceID())
				if err != nil {
					s.logger.Error("reloading state following apply", "error", err, "run", run)
					return
				}
				if _, err := s.tasks.Create(spec); err != nil {
					s.logger.Error("reloading state following apply", "error", err, "run", run)
					return
				}
			}
		},
	}
	return spec, nil
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
