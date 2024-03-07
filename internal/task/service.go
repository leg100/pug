package task

import (
	"context"
	"log/slog"
	"slices"
	"sync"

	"github.com/google/uuid"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"golang.org/x/exp/maps"
)

type Service struct {
	// Tasks keyed by ID
	tasks  map[resource.ID]*Task
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
		tasks:   make(map[resource.ID]*Task),
		factory: factory,
	}

	// Start task enqueuer in background
	enqueuerSub, _ := broker.Subscribe(ctx)
	go startEnqueuer(ctx, svc, enqueuerSub)

	// Start task runner in background
	runner := newRunner(opts.MaxTasks, svc)
	runnerSub, _ := broker.Subscribe(ctx)
	go runner.start(ctx, runnerSub)

	return svc
}

// Create a task. The task is placed into a pending state and requires enqueuing
// before it'll be processed.
func (s *Service) Create(opts CreateOptions) (*Task, error) {
	task, err := s.newTask(opts)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.tasks[task.ID] = task
	s.mu.Unlock()

	if opts.AfterCreate != nil {
		opts.AfterCreate(task)
	}

	go func() {
		if err := task.Wait(); err != nil {
			// TODO: log error
			return
		}
	}()

	s.broker.Publish(resource.CreatedEvent, task)
	slog.Info("created task", "task_id", task.ID, "command", task.Command)
	return task, nil
}

// Enqueue moves the task onto the global queue for processing.
func (s *Service) Enqueue(taskID resource.ID) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[taskID]
	if !ok {
		slog.Error("enqueuing task", "error", "task not found", "task_id", taskID.String())
		return nil, resource.ErrNotFound
	}

	task.updateState(Queued)
	s.broker.Publish(resource.UpdatedEvent, task)
	slog.Info("enqueued task", "task_id", task.ID, "command", task.Command)
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
	// Filter tasks by only those that are block. If false, both blocking and
	// non-blocking tasks are returned.
	Blocking bool
	// Filter tasks by only those that have an ancestor with the given ID.
	Ancestor resource.ID
}

type taskLister interface {
	List(opts ListOptions) []*Task
}

func (s *Service) List(opts ListOptions) []*Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks := maps.Values(s.tasks)
	filtered := make([]*Task, 0, len(tasks))

	for _, t := range tasks {
		if opts.Path != nil && *opts.Path != t.Path {
			// skip tasks matching different path
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
		if opts.Ancestor != resource.ID(uuid.Nil) {
			if !t.HasAncestor(opts.Ancestor) {
				continue
			}
		}
		filtered = append(filtered, t)
	}

	slices.SortFunc(filtered, func(a, b *Task) int {
		cmp := a.Updated.Compare(b.Updated)
		if opts.Oldest {
			return cmp
		}
		return -cmp
	})

	return filtered
}

func (s *Service) Get(taskID resource.ID) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return nil, resource.ErrNotFound
	}
	return task, nil
}

func (s *Service) Subscribe(ctx context.Context) (<-chan resource.Event[*Task], func()) {
	return s.broker.Subscribe(ctx)
}

func (s *Service) Cancel(id resource.ID) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return nil, resource.ErrNotFound
	}
	task.cancel()
	return task, nil
}

func (s *Service) Delete(id resource.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// check task exists, and send it in a deleted event.
	task, ok := s.tasks[id]
	if !ok {
		return resource.ErrNotFound
	}
	// TODO: only allow deleting task if in finished state (error message should
	// instruct user to cancel task first).
	delete(s.tasks, id)
	s.broker.Publish(resource.DeletedEvent, task)
	return nil
}
