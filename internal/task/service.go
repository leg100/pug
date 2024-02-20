package task

import (
	"context"
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
	Program string
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

	// subscribe to task events, update cache, and schedule tasks accordingly.
	go func() {
		cache := &Categories{}
		sub, unsub := broker.Subscribe(ctx)
		defer unsub()

		for event := range sub {
			cache.Categorize(event)
			enqueue := event.Payload.Scheduler.Handle(event)
			for _, t := range enqueue {
				t.updateState(Queued)
			}
		}
	}()

	return &Service{
		broker:  broker,
		factory: factory,
	}
}

// Create a task. The task is placed into a pending state and requires enqueuing
// before it'll be processed.
func (s *Service) Create(spec Spec) (*Task, error) {
	task, err := s.newTask(spec.Path, spec.Args...)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.tasks[task.id] = task
	s.mu.Unlock()

	s.broker.Publish(resource.CreatedEvent, task)
	return task, nil
}

// Enqueue moves the task onto the global queue for processing.
func (s *Service) Enqueue(spec Spec) (*Task, error) {
	task, err := s.newTask(spec.Path, spec.Args...)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.tasks[task.id] = task
	s.mu.Unlock()

	s.enqueue(task, spec.Exclusive)
	s.broker.Publish(resource.CreatedEvent, task)
	return task, nil
}

func (s *Service) List() []*Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	return maps.Values(s.tasks)
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
