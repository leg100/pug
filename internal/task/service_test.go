package task

import (
	"testing"

	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestService_List(t *testing.T) {
	queued := &Task{Resource: resource.New(resource.Task, "", nil), State: Queued}

	tests := []struct {
		name string
		opts ListOptions
		want []*Task
	}{
		{
			"list all",
			ListOptions{},
			[]*Task{queued},
		},
		{
			"list queued",
			ListOptions{Status: []Status{Queued}},
			[]*Task{queued},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				tasks: map[resource.ID]*Task{
					queued.ID: queued,
				},
			}
			assert.Equal(t, tt.want, svc.List(tt.opts))
		})
	}
}
