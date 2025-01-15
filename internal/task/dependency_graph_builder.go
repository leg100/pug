package task

import (
	"fmt"

	"github.com/leg100/pug/internal/resource"
)

type taskCreator interface {
	Create(spec Spec) (*Task, error)
}

// createDependentTasks creates tasks whilst respecting their modules'
// dependencies.
func createDependentTasks(svc taskCreator, reverse bool, specs ...Spec) ([]*Task, error) {
	b := dependencyGraphBuilder{
		nodes:       make(map[resource.ID]*dependencyGraphNode),
		taskCreator: svc,
	}
	// Build dependency graph. Each node in the graph is a module together with
	// the specs that belong to that module. Once the graph is built and
	// dependencies are established, only then are tasks created from the specs.
	for _, spec := range specs {
		node, ok := b.nodes[spec.ModuleID]
		if !ok {
			node = &dependencyGraphNode{dependencies: spec.Dependencies.ModuleIDs}
		}
		node.specs = append(node.specs, spec)
		b.nodes[spec.ModuleID] = node
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

	if len(b.tasks) == 0 {
		return nil, fmt.Errorf("failed to create all %d tasks; see logs", len(b.createErrors))
	}
	return b.tasks, nil
}

// dependencyGraphBuilder builds a graph of dependencies
type dependencyGraphBuilder struct {
	tasks        []*Task
	createErrors []error
	nodes        map[resource.ID]*dependencyGraphNode

	taskCreator
}

// dependencyGraphNode represents a module in a dependency graph
type dependencyGraphNode struct {
	dependencies []resource.ID
	specs        []Spec
	created      []resource.ID
	in, out      []*dependencyGraphNode
	visited      bool
	tasksCreated bool
}

// visit nodes recursively, populating the in and out degrees.
func (b *dependencyGraphBuilder) visit(id resource.ID, n *dependencyGraphNode) {
	n.visited = true

	for _, id := range n.dependencies {
		if dep, ok := b.nodes[id]; ok {
			if !dep.visited {
				b.visit(id, dep)
			}
			dep.in = append(dep.in, n)
			n.out = append(n.out, dep)
		}
	}
}

func (b *dependencyGraphBuilder) visitAndCreateTasks(n *dependencyGraphNode) {
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
		spec.dependsOn = dependsOn
		if task := b.createTask(spec); task != nil {
			n.created = append(n.created, task.ID)
		}
	}
}

func (b *dependencyGraphBuilder) visitAndCreateTasksInReverse(n *dependencyGraphNode) {
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
		spec.dependsOn = dependsOn
		if task := b.createTask(spec); task != nil {
			n.created = append(n.created, task.ID)
		}
	}
}

func (b *dependencyGraphBuilder) createTask(spec Spec) *Task {
	task, err := b.Create(spec)
	if err != nil {
		b.createErrors = append(b.createErrors, err)
	} else {
		b.tasks = append(b.tasks, task)
	}
	return task
}
