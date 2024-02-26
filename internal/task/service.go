package task

import (
	"context"
	"slices"
	"sync"

	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"golang.org/x/exp/maps"
)

type Service struct {
	// Tasks keyed by ID
	tasks  map[resource.Resource]*Task
	broker *pubsub.Broker[*Task]
	// Mutex for concurrent read/write of tasks
	mu sync.Mutex

	*factory
}

type ServiceOptions struct {
	MaxTasks int
	Program  string
}

func NewService(ctx context.Context, opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*Task]()
	factory := &factory{
		program: opts.Program,
		// Publish an event whenever task state is updated
		afterUpdate: func(t *Task) {
			broker.Publish(resource.UpdatedEvent, t)
		},
	}

	svc := &Service{
		broker:  broker,
		factory: factory,
	}

	// Start task runner in background
	runner := newRunner(opts.MaxTasks, svc)
	go runner.start(ctx)

	return svc
}

type CreateOptions struct {
	// Resource that the task belongs to.
	Parent resource.Resource
	// Program command and any sub commands, e.g. plan, state rm, etc.
	Command []string
	// Args to pass to program.
	Args []string
	// Environment variables.
	Env []string
	// Path in which to execute the program - assumed be the terraform module's
	// path.
	Path string
	// Arbitrary metadata to associate with the task.
	Metadata map[string]string
	// A blocking task blocks other tasks from running on the module or
	// workspace.
	Blocking bool
	// Globally exclusive task - at most only one such task can be running
	Exclusive bool
	// Call this function after the task has successfully finished
	AfterExited func(*Task)
	// Call this function after the task fails with an error
	AfterError func(*Task)
	// Call this function after the task is successfully created
	AfterCreate func(*Task)
}

// Create a task. The task is placed into a pending state and requires enqueuing
// before it'll be processed.
func (s *Service) Create(opts CreateOptions) (*Task, error) {
	task, err := s.newTask(opts)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.tasks[task.Resource] = task
	s.mu.Unlock()

	if opts.AfterCreate != nil {
		opts.AfterCreate(task)
	}

	go func() {
		if err := task.Wait(); err != nil {
			// TODO: move this into task itself, before an update event is
			// published.
			if opts.AfterError != nil {
				opts.AfterError(task)
			}
			// TODO: log error
			return
		}
		// TODO: move this into task itself, before an update event is
		// published.
		if opts.AfterExited != nil {
			opts.AfterExited(task)
		}
	}()

	s.broker.Publish(resource.CreatedEvent, task)
	return task, nil
}

// Enqueue moves the task onto the global queue for processing.
func (s *Service) Enqueue(id resource.Resource) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return nil, resource.ErrNotFound
	}

	task.updateState(Queued)
	s.broker.Publish(resource.UpdatedEvent, task)
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
}

func (s *Service) List(opts ListOptions) []*Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	var filtered []*Task
	for _, t := range maps.Values(s.tasks) {
		if opts.Path != nil && *opts.Path != t.Path {
			// skip tasks matching different path
			continue
		}
		if opts.Status != nil {
			if !slices.Contains(opts.Status, t.State) {
				continue
			}
		}
		filtered = append(filtered, t)
	}
	slices.SortFunc(filtered, func(a, b *Task) int {
		cmp := a.updated.Compare(b.updated)
		if opts.Oldest {
			return cmp
		}
		return -cmp
	})

	return filtered
}

func (s *Service) Watch(ctx context.Context) (<-chan resource.Event[*Task], func()) {
	return s.broker.Subscribe(ctx)
}

func (s *Service) Cancel(id resource.Resource) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return nil, resource.ErrNotFound
	}
	task.cancel()
	return task, nil
}

func (s *Service) Delete(id resource.Resource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// check task exists, and send it in a deleted event.
	task, ok := s.tasks[id]
	if !ok {
		return resource.ErrNotFound
	}
	delete(s.tasks, id)
	s.broker.Publish(resource.DeletedEvent, task)
	return nil
}
