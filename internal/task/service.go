package task

import (
	"slices"

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
	task, err := s.newTask(spec)
	if err != nil {
		return nil, err
	}

	s.logger.Info("created task", "task", task)

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
		s.logger.Info("completed task", "task", task)
	}()
	if spec.Wait {
		return task, <-wait
	}
	return task, nil
}

// Create a task group from one or more task specs. An error is returned if zero
// specs are provided, or if it fails to create at least one task.
func (s *Service) CreateGroup(specs ...Spec) (*Group, error) {
	g, err := newGroup(s, specs...)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("created task group", "group", g)

	// Add to db
	s.AddGroup(g)

	return g, nil
}

// AddGroup adds a task group to the DB.
func (s *Service) AddGroup(group *Group) {
	s.groups.Add(group.ID, group)
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
