package task

import (
	"github.com/google/uuid"
	"github.com/leg100/pug/internal/resource"
)

// Categories maintains a set of categories of tasks to assist scheduling and
// presentation to users.
type Categories struct {
	// * pending (oldest first)
	Pending []*Task
	// * queued (oldest first)
	Queued []*Task
	// * running (newest first)
	Running []*Task
	// * finished (newest first)
	Finished []*Task
}

func (s *Categories) Active() []*Task {
	return append(s.Queued, s.Running...)
}

func (s *Categories) Categorize(event resource.Event[*Task]) {
	switch event.Type {
	case resource.CreatedEvent:
		// Newly created tasks should always be in a pending state
		s.Pending = append(s.Pending, event.Payload)
	case resource.UpdatedEvent:
		// Move event from one category to another.
		// First remove task from whatever category it is in.
		s.removeTask(event.Payload.id)
		// Then add to appropriate category based on updated state
		switch event.Payload.State {
		case Pending:
			// should not happen
		case Queued:
			// Add to end of queued category
			s.Queued = append(s.Queued, event.Payload)
		case Running:
			// Add to start of running category
			s.Running = append([]*Task{event.Payload}, s.Running...)
		default:
			// Add to start of finished category
			s.Finished = append([]*Task{event.Payload}, s.Finished...)
		}
	case resource.DeletedEvent:
		s.removeTask(event.Payload.id)
	}
}

func (s *Categories) removeTask(id uuid.UUID) {
	for i, t := range s.Pending {
		if id == t.id {
			s.Pending = append(s.Pending[:i], s.Pending[i+1:]...)
			return
		}
	}
	for i, t := range s.Queued {
		if id == t.id {
			s.Queued = append(s.Queued[:i], s.Queued[i+1:]...)
			return
		}
	}
	for i, t := range s.Running {
		if id == t.id {
			s.Running = append(s.Running[:i], s.Running[i+1:]...)
			return
		}
	}
	for i, t := range s.Finished {
		if id == t.id {
			s.Finished = append(s.Finished[:i], s.Finished[i+1:]...)
			return
		}
	}
}
