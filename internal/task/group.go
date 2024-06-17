package task

import (
	"errors"
	"slices"
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
		Command: cmd,
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

func (g *Group) IncludesTask(taskID resource.ID) bool {
	return slices.ContainsFunc(g.Tasks, func(tgt *Task) bool {
		return tgt.ID == taskID
	})
}

func (g *Group) Finished() int {
	var finished int
	for _, t := range g.Tasks {
		if t.IsFinished() {
			finished++
		}
	}
	return finished
}

func (g *Group) Exited() int {
	var exited int
	for _, t := range g.Tasks {
		if t.State == Exited {
			exited++
		}
	}
	return exited
}

func (g *Group) Errored() int {
	var errored int
	for _, t := range g.Tasks {
		if t.State == Errored {
			errored++
		}
	}
	return errored
}

func SortGroupsByCreated(i, j *Group) int {
	if i.Created.After(j.Created) {
		return -1
	}
	return 1
}
