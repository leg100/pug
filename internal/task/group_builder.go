package task

import (
	"fmt"

	"github.com/leg100/pug/internal/resource"
)

type taskCreator interface {
	Create(spec CreateOptions) (*Task, error)
}

// NewGroupWithDependencies constructs a graph from the given task specs.
func NewGroupWithDependencies(svc taskCreator, cmd string, specs ...CreateOptions) (*Group, error) {
	b := groupBuilder{
		g:     NewEmptyGroup(cmd),
		nodes: make(map[resource.ID]*groupBuilderNode),
		svc:   svc,
	}
	// Enumerate thru specs, check if spec belongs to a module, if so then
	// create a node if one doesn't exist already and add spec to node. If spec
	// does not belong to a module then create the task now (because it doesn't
	// have any dependencies).
	for _, spec := range specs {
		if mod := spec.Parent.Module(); mod != nil {
			modID := mod.GetID()
			node, ok := b.nodes[modID]
			if !ok {
				node = &groupBuilderNode{module: mod}
			}
			node.specs = append(node.specs, spec)
			b.nodes[modID] = node
		} else {
			b.createTask(spec)
		}
	}

	for id, v := range b.nodes {
		if !v.visited {
			b.visit(id, v)
		}
	}
	if len(b.g.Tasks) == 0 {
		return b.g, fmt.Errorf("failed to create all %d tasks; see logs", len(b.g.CreateErrors))
	}
	return b.g, nil
}

// groupBuilder builds a task group.
type groupBuilder struct {
	g     *Group
	svc   taskCreator
	nodes map[resource.ID]*groupBuilderNode
}

// groupBuilderNode represents a group of task specs that belong to the same
// module.
type groupBuilderNode struct {
	module  resource.Resource
	specs   []CreateOptions
	created []resource.ID
	visited bool
}

// visit traverses the node dependency tree starting at u.
func (b *groupBuilder) visit(id resource.ID, u *groupBuilderNode) {
	u.visited = true
	var taskDependencies []resource.ID
	for _, id := range u.module.Dependencies() {
		if v, ok := b.nodes[id]; ok {
			if !v.visited {
				b.visit(id, v)
			}
			taskDependencies = append(taskDependencies, v.created...)
		}
	}
	// For each spec, add dependencies on other tasks before creating task and
	// adding its ID to the vertex
	for _, spec := range u.specs {
		spec.DependsOn = taskDependencies
		if task := b.createTask(spec); task != nil {
			u.created = append(u.created, task.ID)
		}
	}
}

func (b *groupBuilder) createTask(spec CreateOptions) *Task {
	task, err := b.svc.Create(spec)
	if err != nil {
		b.g.CreateErrors = append(b.g.CreateErrors, err)
	} else {
		b.g.Tasks = append(b.g.Tasks, task)
	}
	return task
}
