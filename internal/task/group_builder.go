package task

import (
	"fmt"

	"github.com/leg100/pug/internal/resource"
)

type taskCreator interface {
	Create(spec CreateOptions) (*Task, error)
}

// newGroupWithDependencies constructs a graph from the given task specs.
func newGroupWithDependencies(svc taskCreator, cmd string, reverse bool, specs ...CreateOptions) (*Group, error) {
	b := groupBuilder{
		g:     NewEmptyGroup(cmd),
		nodes: make(map[resource.ID]*groupBuilderNode),
		svc:   svc,
	}
	// Build dependency graph. Each node in the graph is a module and the specs
	// that belong to that module. Once the graph is built and dependencies are
	// established, only then are tasks created from the specs.
	//
	// Specs that don't belong to a module don't have any dependencies so tasks
	// are built from these specs immediately.
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
	// Now create tasks. If reverse is true, then create tasks in reverse order
	// to dependencies, e.g. where a module A depends on module B, create tasks
	// on module A before module B. This is necessary when the tasks are
	// destroying infrastructure using `terraform apply -destroy`.
	for _, n := range b.nodes {
		if !n.tasksCreated {
			if reverse {
				b.visitAndCreateTasksInReverse(n)
			} else {
				b.visitAndCreateTasks(n)
			}
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
	module       resource.Resource
	specs        []CreateOptions
	created      []resource.ID
	in, out      []*groupBuilderNode
	visited      bool
	tasksCreated bool
}

// visit nodes recursively, populating the in and out degrees.
func (b *groupBuilder) visit(id resource.ID, n *groupBuilderNode) {
	n.visited = true

	for _, id := range n.module.Dependencies() {
		if dep, ok := b.nodes[id]; ok {
			if !dep.visited {
				b.visit(id, dep)
			}
			dep.in = append(dep.in, n)
			n.out = append(n.out, dep)
		}
	}
}

func (b *groupBuilder) visitAndCreateTasks(n *groupBuilderNode) {
	n.tasksCreated = true

	var dependsOn []resource.ID
	for _, out := range n.out {
		if !out.tasksCreated {
			b.visitAndCreateTasks(out)
		}
		dependsOn = append(dependsOn, out.created...)
	}
	// For each spec, add dependencies on other tasks before creating task and
	// adding its ID to the node
	for _, spec := range n.specs {
		spec.DependsOn = dependsOn
		if task := b.createTask(spec); task != nil {
			n.created = append(n.created, task.ID)
		}
	}
}

func (b *groupBuilder) visitAndCreateTasksInReverse(n *groupBuilderNode) {
	n.tasksCreated = true

	var dependsOn []resource.ID
	for _, in := range n.in {
		if !in.tasksCreated {
			b.visitAndCreateTasksInReverse(in)
		}
		dependsOn = append(dependsOn, in.created...)
	}
	// For each spec, add dependencies on other tasks before creating task and
	// adding its ID to the node
	for _, spec := range n.specs {
		spec.DependsOn = dependsOn
		if task := b.createTask(spec); task != nil {
			n.created = append(n.created, task.ID)
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
