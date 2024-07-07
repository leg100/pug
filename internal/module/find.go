package module

import (
	"io"
	"strings"

	"github.com/awalterschulze/gographviz"
)

// parseTerragruntGraph parses the output from `terragrunt graph-dependencies`
// and returns a graph of module paths and their dependencies on one another.
//
// TODO: `terragrunt graph-dependencies` can return "external dependencies" i.e.
// those outside of pug's working directory, which we want to exclude.
func parseTerragruntGraph(r io.Reader) (map[string][]string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	// Parse graphviz output and add dependency info to each module.
	astgraph, err := gographviz.Parse(b)
	if err != nil {
		return nil, err
	}
	graph, err := gographviz.NewAnalysedGraph(astgraph)
	if err != nil {
		return nil, err
	}
	// First populate map with nodes representing each module path
	m := make(map[string][]string, len(graph.Nodes.Nodes))
	for _, node := range graph.Nodes.Nodes {
		m[stripDoubleQuotes(node.Name)] = nil
	}
	// Next, add dependencies, represented by edges, to each module in the map.
	for _, e := range graph.Edges.Edges {
		path := stripDoubleQuotes(e.Src)
		dep := stripDoubleQuotes(e.Dst)
		m[path] = append(m[path], dep)
	}
	return m, nil
}

// terragrunt graph nodes and edges, for some reason, have embedded double
// quotes, so we need to strip them.
func stripDoubleQuotes(s string) string {
	return strings.ReplaceAll(s, `"`, "")
}
