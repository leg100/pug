package task

import (
	"errors"
	"time"

	"github.com/leg100/pug/internal/resource"
)

type Group struct {
	resource.Common

	Created      time.Time
	Command      string
	Tasks        []*Task
	CreateErrors []error
}

func newGroup(cmd string, fn Func, ids ...resource.ID) (*Group, error) {
	g := &Group{
		Common:  resource.New(resource.TaskGroup, resource.GlobalResource),
		Created: time.Now(),
	}
	for _, id := range ids {
		task, err := fn(id)
		if err != nil {
			g.CreateErrors = append(g.CreateErrors, err)
		} else {
			g.Tasks = append(g.Tasks, task)
		}
	}
	// If no tasks were created, then return error.
	if len(g.Tasks) == 0 {
		return nil, errors.New("all tasks failed to be created")
	}
	return g, nil
}

func (g *Group) String() string { return g.Command }

func SortGroupsByCreated(i, j *Group) int {
	if i.Created.Before(j.Created) {
		return -1
	}
	return 1
}
