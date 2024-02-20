package resource

const (
	CreatedEvent EventType = "created"
	UpdatedEvent EventType = "updated"
	DeletedEvent EventType = "deleted"
)

type (
	// EventType identifies the type of event
	EventType string

	// Event represents an event in the lifecycle of a resource
	Event[T any] struct {
		Type    EventType
		Payload T
	}
)

func NewEvent[T any](t EventType, payload T) Event[T] {
	return Event[T]{Type: t, Payload: payload}
}
