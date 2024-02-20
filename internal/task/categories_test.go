package task

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestCategories(t *testing.T) {
	create := func(t *Task) resource.Event[*Task] {
		return resource.Event[*Task]{
			Payload: t,
			Type:    resource.CreatedEvent,
		}
	}
	update := func(t *Task) resource.Event[*Task] {
		return resource.Event[*Task]{
			Payload: t,
			Type:    resource.UpdatedEvent,
		}
	}
	deletefn := func(t *Task) resource.Event[*Task] {
		return resource.Event[*Task]{
			Payload: t,
			Type:    resource.DeletedEvent,
		}
	}

	task1 := &Task{id: uuid.New()}
	task2 := &Task{id: uuid.New()}
	task3 := &Task{id: uuid.New()}
	cats := &Categories{}

	cats.Categorize(create(task1))
	want := &Categories{Pending: []*Task{task1}}
	assert.Equal(t, want, cats)

	cats.Categorize(create(task2))
	want = &Categories{Pending: []*Task{task1, task2}}
	assert.Equal(t, want, cats)

	cats.Categorize(create(task3))
	want = &Categories{Pending: []*Task{task1, task2, task3}}
	assert.Equal(t, want, cats)

	task1.State = Queued
	cats.Categorize(update(task1))
	want = &Categories{
		Pending: []*Task{task2, task3},
		Queued:  []*Task{task1},
	}
	assert.Equal(t, want, cats)

	task1.State = Running
	cats.Categorize(update(task1))
	want = &Categories{
		Pending: []*Task{task2, task3},
		Queued:  []*Task{},
		Running: []*Task{task1},
	}
	assert.Equal(t, want, cats)

	cats.Categorize(deletefn(task3))
	want = &Categories{
		Pending: []*Task{task2},
		Queued:  []*Task{},
		Running: []*Task{task1},
	}
	assert.Equal(t, want, cats)

	task1.State = Exited
	cats.Categorize(update(task1))
	want = &Categories{
		Pending:  []*Task{task2},
		Queued:   []*Task{},
		Running:  []*Task{},
		Finished: []*Task{task1},
	}
	assert.Equal(t, want, cats)
}
