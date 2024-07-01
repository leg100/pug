package module

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/awalterschulze/gographviz"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"golang.org/x/exp/maps"
)

// findResult is the result of successfully finding a module on the filesystem.
type findResult struct {
	path         string
	dependencies []string
}

// findModules finds root modules that are descendents of the workdir and
// returns their paths relative to the workdir.
//
// A root module is deemed to be a directory that contains a .tf file that
// contains a backend block.
func findModules(logger logging.Interface, workdir internal.Workdir) ([]findResult, error) {
	found := make(map[string]struct{})
	walkfn := func(path string, d fs.DirEntry, walkerr error) error {
		// skip directories that have already been identified as containing a
		// root module
		if _, ok := found[filepath.Dir(path)]; ok {
			return nil
		}
		if walkerr != nil {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case ".terraform", ".terragrunt-cache":
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "terragrunt.hcl" {
			found[filepath.Dir(path)] = struct{}{}
			// skip walking remainder of parent directory
			return fs.SkipDir
		}
		if filepath.Ext(path) != ".tf" {
			return nil
		}
		cfg, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		// only the hclwrite pkg seems to have the ability to walk HCL blocks,
		// so this is what is used even though no writing is performed.
		file, diags := hclwrite.ParseConfig(cfg, path, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			logger.Error("reloading modules: parsing hcl", "path", path, "error", diags)
			return nil
		}
		for _, block := range file.Body().Blocks() {
			if block.Type() == "terraform" {
				for _, nested := range block.Body().Blocks() {
					if nested.Type() == "backend" || nested.Type() == "cloud" {
						found[filepath.Dir(path)] = struct{}{}
						// skip walking remainder of parent directory
						return fs.SkipDir
					}
				}
			}
		}
		return nil
	}
	if err := filepath.WalkDir(workdir.String(), walkfn); err != nil {
		return nil, err
	}
	// Strip parent prefix from paths before returning
	results := make([]findResult, len(found))
	for i, f := range maps.Keys(found) {
		stripped, err := filepath.Rel(workdir.String(), f)
		if err != nil {
			return nil, err
		}
		results[i] = findResult{path: stripped}
	}
	return results, nil
}

func findTerragruntModules(graphOutput io.Reader) ([]findResult, error) {
	b, err := io.ReadAll(graphOutput)
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
	mods := make(map[string][]string)
	for _, node := range graph.Nodes.Nodes {
		node.Attrs.Add
	for _, e := range graph.Edges.Edges {
		// strip embedded double quotes from module name
		name := stripDoubleQuotes(e.Src)
		// if module not found before, then allocate slice for its dependencies
		deps, ok := mods[name]
		if !ok {
			deps = make([]string, 0)
		}
		// append dependency (after stripping embedded double quotes)
		mods[name] = append(deps, stripDoubleQuotes(e.Dst))
	}
	// Re-structure map into []FindResult
	results := make([]findResult, len(mods))
	var i int
	for path, deps := range mods {
		results[i] = findResult{path: path, dependencies: deps}
		i++
	}
	return results, nil
}

func stripDoubleQuotes(s string) string {
	return strings.ReplaceAll(s, `"`, "")
}
