package task

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
)

type Service struct {
	tasks   *resource.Table[*Task]
	groups  *resource.Table[*Group]
	counter *int
	logger  logging.Interface

	TaskBroker  *pubsub.Broker[*Task]
	GroupBroker *pubsub.Broker[*Group]
	*factory
}

type ServiceOptions struct {
	Program    string
	Logger     logging.Interface
	Workdir    internal.Workdir
	UserEnvs   []string
	UserArgs   []string
	Terragrunt bool
}

func NewService(opts ServiceOptions) *Service {
	var counter int

	taskBroker := pubsub.NewBroker[*Task](opts.Logger)
	groupBroker := pubsub.NewBroker[*Group](opts.Logger)

	factory := &factory{
		publisher:  taskBroker,
		counter:    &counter,
		program:    opts.Program,
		workdir:    opts.Workdir,
		userEnvs:   opts.UserEnvs,
		userArgs:   opts.UserArgs,
		terragrunt: opts.Terragrunt,
	}

	return &Service{
		tasks:       resource.NewTable(taskBroker),
		groups:      resource.NewTable(groupBroker),
		TaskBroker:  taskBroker,
		GroupBroker: groupBroker,
		factory:     factory,
		counter:     &counter,
		logger:      opts.Logger,
	}
}

// Create a task. The task is placed into a pending state and requires enqueuing
// before it'll be processed.
func (s *Service) Create(spec Spec) (*Task, error) {
	task := s.newTask(spec)

	s.logger.Debug("created task", "task", task)

	// Add to db
	s.tasks.Add(task.ID, task)
	// Increment counter of number of live tasks
	*s.counter++

	if spec.AfterCreate != nil {
		spec.AfterCreate(task)
	}

	wait := make(chan error, 1)
	go func() {
		err := task.Wait()
		wait <- err
		if err != nil {
			s.logger.Error("task failed", "error", err, "task", task)
			return
		}
		s.logger.Debug("completed task", "task", task)
	}()
	if spec.Wait {
		return task, <-wait
	}
	return task, nil
}

// CreateGroup creates a task group, creating tasks by invoking the provided
// func with each of the provided IDs.
func (s *Service) CreateGroup(cmd string, fn Func, ids ...resource.ID) (*Group, error) {
	group, err := newGroup(cmd, fn, ids...)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("created task group", "group", group)

	// Add to db
	s.AddGroup(group)

	return group, nil
}

func (s *Service) CreateDependencyGroup(cmd string, reverse bool, opts ...Spec) (*Group, error) {
	if len(opts) == 0 {
		return nil, errors.New("no specs provided")
	}

	group, err := newGroupWithDependencies(s, cmd, reverse, opts...)
	if err != nil {
		return nil, err
	}
	s.logger.Debug("created task group", "group", group)

	// Add to db
	s.AddGroup(group)

	return group, nil
}

// SpecErrors are errors returned from attempting to create task specs.
type SpecErrors []error

func (e SpecErrors) Error() string {
	return fmt.Sprintf("failed to create %d specs", len(e))
}

// Create a task group from one or more task spec functions. An error is returned if zero
// functions are provided, or if it fails to create at least one task.
func (s *Service) CreateGroupFromSpecs(funcs ...SpecFunc) (*Group, error) {
	if len(funcs) == 0 {
		return nil, errors.New("no spec functions provided")
	}
	g := &Group{
		Common:  resource.New(resource.TaskGroup, resource.GlobalResource),
		Created: time.Now(),
	}
	var (
		cmd    string
		pruned SpecErrors
	)
	for _, fn := range funcs {
		spec, err := fn()
		if err != nil {
			pruned = append(pruned, err)
		}
		task, err := s.Create(spec)
		if err != nil {
			g.CreateErrors = append(g.CreateErrors, err)
		}
		if cmd == "" {
			cmd = task.String()
		} else if cmd != task.String() {
			// TODO: make a constant
			cmd = "multi"
		}
		g.Tasks = append(g.Tasks, task)
	}
	if pruned != nil {
		return nil, pruned
	}
	if len(g.Tasks) == 0 {
		return g, errors.New("all tasks failed to be created")
	}
	return g, nil
}

// AddGroup adds a task group to the DB.
func (s *Service) AddGroup(group *Group) {
	s.groups.Add(group.ID, group)
}

// Retry creates a new task that has all the properties of the task with the
// given ID.
func (s *Service) Retry(taskID resource.ID) (*Task, error) {
	task, err := s.Get(taskID)
	if err != nil {
		return nil, err
	}
	return s.Create(task.createOptions)
}

// Enqueue moves the task onto the global queue for processing.
func (s *Service) Enqueue(taskID resource.ID) (*Task, error) {
	task, err := s.tasks.Update(taskID, func(existing *Task) error {
		existing.updateState(Queued)
		return nil
	})
	if err != nil {
		s.logger.Error("enqueuing task", "error", err)
		return nil, err
	}
	s.logger.Debug("enqueued task", "task", task)
	return task, nil
}

type ListOptions struct {
	// Filter tasks by those with a matching module path. Optional.
	Path *string
	// Filter tasks by status: match task if it has one of these statuses.
	// Optional.
	Status []Status
	// Order tasks by oldest first (true), or newest first (false)
	Oldest bool
	// Filter tasks by only those that are blocking. If false, both blocking and
	// non-blocking tasks are returned.
	Blocking bool
	// Only return those tasks that are exclusive. If false, both exclusive and
	// non-exclusive tasks are returned.
	Exclusive bool
	// Filter tasks by those with one of the following commands
	Command [][]string
	// Filter tasks by only those that have an ancestor with the given ID.
	// Defaults the zero value, which is the ID of the abstract global entity to
	// which all resources belong.
	Ancestor resource.ID
}

type taskLister interface {
	List(opts ListOptions) []*Task
}

func (s *Service) List(opts ListOptions) []*Task {
	tasks := s.tasks.List()

	// Filter list according to options
	var i int
	for _, t := range tasks {
		if opts.Path != nil && *opts.Path != t.Path {
			continue
		}
		if opts.Status != nil {
			if !slices.Contains(opts.Status, t.State) {
				continue
			}
		}
		if opts.Blocking {
			if !t.Blocking {
				continue
			}
		}
		if opts.Exclusive {
			if !t.exclusive {
				continue
			}
		}
		if opts.Command != nil {
			for _, cmd := range opts.Command {
				if slices.Equal(cmd, t.Command) {
					break
				}
			}
		}
		if !t.HasAncestor(opts.Ancestor) {
			continue
		}
		tasks[i] = t
		i++
	}
	tasks = tasks[:i]

	// Sort list according to options
	slices.SortFunc(tasks, func(a, b *Task) int {
		cmp := a.Updated.Compare(b.Updated)
		if opts.Oldest {
			return cmp
		}
		return -cmp
	})

	return tasks
}

func (s *Service) ListGroups() []*Group {
	return s.groups.List()
}

func (s *Service) Get(taskID resource.ID) (*Task, error) {
	return s.tasks.Get(taskID)
}

func (s *Service) GetGroup(groupID resource.ID) (*Group, error) {
	return s.groups.Get(groupID)
}

func (s *Service) Cancel(taskID resource.ID) (*Task, error) {
	task, err := func() (*Task, error) {
		task, err := s.tasks.Get(taskID)
		if err != nil {
			return nil, err
		}
		return task, task.cancel()
	}()
	if err != nil {
		s.logger.Error("canceling task", "id", taskID, "error", err)
		return nil, err
	}

	s.logger.Info("canceled task", "task", task)
	return task, nil
}

func (s *Service) Delete(taskID resource.ID) error {
	// TODO: only allow deleting task if in finished state (error message should
	// instruct user to cancel task first).
	s.tasks.Delete(taskID)
	return nil
}

func (s *Service) Counter() int {
	return *s.counter
}
