package task

import (
	"testing"

	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestService_List(t *testing.T) {
	t.Parallel()

	pending := &Task{ID: resource.NewID(resource.Task), State: Pending}
	queued := &Task{ID: resource.NewID(resource.Task), State: Queued}
	running := &Task{ID: resource.NewID(resource.Task), State: Running}
	exited := &Task{ID: resource.NewID(resource.Task), State: Exited}
	errored := &Task{ID: resource.NewID(resource.Task), State: Errored}

	tests := []struct {
		name string
		opts ListOptions
		want func(t *testing.T, got []*Task)
	}{
		{
			"list all",
			ListOptions{},
			func(t *testing.T, got []*Task) {
				assert.Equal(t, 5, len(got))
			},
		},
		{
			"list pending",
			ListOptions{Status: []Status{Pending}},
			func(t *testing.T, got []*Task) {
				assert.Equal(t, 1, len(got))
			},
		},
		{
			"list queued",
			ListOptions{Status: []Status{Queued}},
			func(t *testing.T, got []*Task) {
				assert.Equal(t, 1, len(got))
			},
		},
		{
			"list running",
			ListOptions{Status: []Status{Running}},
			func(t *testing.T, got []*Task) {
				if assert.Equal(t, 1, len(got)) {
					assert.Equal(t, got[0], running)
				}
			},
		},
		{
			"list finished",
			ListOptions{Status: []Status{Exited, Errored}},
			func(t *testing.T, got []*Task) {
				assert.Equal(t, 2, len(got))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// bootstrap service with tasks
			svc := &Service{
				tasks: resource.NewTable(&fakePublisher[*Task]{}),
			}
			svc.tasks.Add(pending.ID, pending)
			svc.tasks.Add(queued.ID, queued)
			svc.tasks.Add(running.ID, running)
			svc.tasks.Add(exited.ID, exited)
			svc.tasks.Add(errored.ID, errored)

			tt.want(t, svc.List(tt.opts))
		})
	}
}
