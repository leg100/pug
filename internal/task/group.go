package task

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/leg100/pug/internal/resource"
)

type Group struct {
	resource.ID

	Created      time.Time
	Command      string
	Tasks        []*Task
	CreateErrors []error
}

func newGroup(service *Service, specs ...Spec) (*Group, error) {
	if len(specs) == 0 {
		return nil, errors.New("no specs provided")
	}
	g := &Group{
		ID:      resource.NewID(resource.TaskGroup),
		Created: time.Now(),
	}
	// Validate specifications. There are some settings that are incompatible
	// with one another within a task group.
	var (
		respectModuleDependencies *bool
		inverseDependencyOrder    *bool
	)
	for _, spec := range specs {
		// All specs must specify Dependencies or not specify Dependencies.
		deps := (spec.Dependencies != nil)
		if respectModuleDependencies == nil {
			respectModuleDependencies = &deps
		} else if *respectModuleDependencies != deps {
			return nil, fmt.Errorf("not all specs share same respect-module-dependencies setting")
		}
		// All specs specifying dependencies must set InverseDependencyOrder to
		// the same value
		inverse := (spec.Dependencies != nil && spec.Dependencies.InverseDependencyOrder)
		if inverseDependencyOrder == nil {
			inverseDependencyOrder = &inverse
		} else if *inverseDependencyOrder != inverse {
			return nil, fmt.Errorf("not all specs share same inverse-dependency-order setting")
		}
	}
	if *respectModuleDependencies {
		tasks, err := createDependentTasks(service, *inverseDependencyOrder, specs...)
		if err != nil {
			return nil, err
		}
		g.Tasks = tasks
	} else {
		for _, spec := range specs {
			task, err := service.Create(spec)
			if err != nil {
				g.CreateErrors = append(g.CreateErrors, err)
				continue
			}
			g.Tasks = append(g.Tasks, task)
		}
	}
	if len(g.Tasks) == 0 {
		return g, errors.New("all tasks failed to be created")
	}

	for _, task := range g.Tasks {
		if g.Command == "" {
			g.Command = task.String()
		} else if g.Command != task.String() {
			// Detected that not all tasks have the same command, so name the
			// task group to reflect that multiple commands comprise the group.
			//
			// TODO: make a constant
			g.Command = "multi"
		}
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
		if t.State.IsFinal() {
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
