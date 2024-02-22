package task

import (
	"context"
	"slices"
	"sync"

	"github.com/google/uuid"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"golang.org/x/exp/maps"
)

type Service struct {
	// Tasks keyed by ID
	tasks  map[uuid.UUID]*Task
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
		// Whenever task state is updated, publish it as an event.
		callback: func(t *Task) {
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

// Create a task. The task is placed into a pending state and requires enqueuing
// before it'll be processed.
func (s *Service) Create(spec Spec) (*Task, error) {
	task, err := s.newTask(spec.Path, spec.Args...)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.tasks[task.ID] = task
	s.mu.Unlock()

	s.broker.Publish(resource.CreatedEvent, task)
	return task, nil
}

// Enqueue moves the task onto the global queue for processing.
func (s *Service) Enqueue(id uuid.UUID) (*Task, error) {
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
	// Filter tasks by those with a matching workspace name. Optional.
	WorkspaceName *string
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
		if opts.WorkspaceName != nil {
			if t.WorkspaceName == nil {
				// skip tasks unassociated with a workspace
				continue
			}
			if *opts.WorkspaceName != *t.WorkspaceName {
				// skip tasks associated with different workspace
				continue
			}
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

func (s *Service) Cancel(id uuid.UUID) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return nil, resource.ErrNotFound
	}
	task.cancel()
	return task, nil
}

func (s *Service) Delete(id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// check task exists, and send it in a deleted event.
	task, ok := s.tasks[id]
	if !ok {
		return resource.ErrNotFound
	}
	s.broker.Publish(resource.DeletedEvent, task)

	delete(s.tasks, id)
	return nil
}
