package run

import (
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
)

// moduleAndResource is a module and the ID of a resource that belongs to the module.
type moduleAndResource struct {
	module *module.Module
	id     resource.ID
}

// graph represents module dependencies.
type graph struct {
	vertices map[resource.ID]*vertex

	// results of the topological sort, ordering the children according to their
	// respective module's dependencies.
	results [][]resource.ID
}

// vertex is a graph vertex, representing a module and the resources that belong
// to that module.
type vertex struct {
	module   *module.Module
	children []resource.ID
	state    vertexState
}

// vertexState is the current state of the vertex.
type vertexState int

const (
	// White means the vertex is unvisited; grey it has been discovered; and
	// black that it is finished. Names taken from Introduction to Algorithms,
	// 3rd edition, p604.
	white vertexState = iota
	grey
	black
)

func newGraph(targets ...moduleAndResource) *graph {
	vertices := make(map[resource.ID]*vertex)
	for _, t := range targets {
		v, ok := vertices[t.module.ID]
		if !ok {
			v = &vertex{module: t.module}
		}
		v.children = append(v.children, t.id)
		vertices[t.module.ID] = v
	}
	return &graph{vertices: vertices}
}

func (g *graph) sort() {
	for _, v := range g.vertices {
		if v.state == white {
			g.visit(v)
		}
	}
}

func (g *graph) visit(u *vertex) {
	u.state = grey
	for _, v := range u.module.Dependencies {
		if v, ok := g.vertices[v.ID]; ok {
			if v.state == white {
				g.visit(v)
			}
		}
	}
	u.state = black
	g.results = append(g.results, u.children)
}
