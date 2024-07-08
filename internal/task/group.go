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

// Func is a function that creates a task.
type Func func(resource.ID) (*Task, error)

// newGroup creates a task group, invoking the provided function on each id to
// each task. If the task is successfully created it is added to the group;
// otherwise the error is added to the group.
func newGroup(cmd string, fn Func, ids ...resource.ID) (*Group, error) {
	g := NewEmptyGroup(cmd)
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

func NewEmptyGroup(cmd string) *Group {
	return &Group{
		Common:  resource.New(resource.TaskGroup, resource.GlobalResource),
		Created: time.Now(),
		Command: cmd,
	}
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
