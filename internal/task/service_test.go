package task

import (
	"testing"

	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestService_List(t *testing.T) {
	t.Parallel()

	mod1 := resource.New(resource.Module, resource.GlobalResource)

	pending := &Task{Common: resource.New(resource.Task, mod1), State: Pending}
	queued := &Task{Common: resource.New(resource.Task, mod1), State: Queued}
	running := &Task{Common: resource.New(resource.Task, mod1), State: Running}
	exited := &Task{Common: resource.New(resource.Task, mod1), State: Exited}
	errored := &Task{Common: resource.New(resource.Task, mod1), State: Errored}

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
				table: resource.NewTable(&fakePublisher[*Task]{}),
			}
			svc.table.Add(pending.ID, pending)
			svc.table.Add(queued.ID, queued)
			svc.table.Add(running.ID, running)
			svc.table.Add(exited.ID, exited)
			svc.table.Add(errored.ID, errored)

			tt.want(t, svc.List(tt.opts))
		})
	}
}
