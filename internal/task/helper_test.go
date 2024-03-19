package task

import "github.com/leg100/pug/internal/resource"

type fakePublisher[T any] struct{}

func (f *fakePublisher[T]) Publish(resource.EventType, T) {}
