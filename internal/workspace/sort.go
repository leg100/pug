package workspace

import (
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
)

type moduleGetter interface {
	Get(id resource.ID) (*module.Module, error)
}

// Sort sorts workspaces accordingly:
//
// 1. first by their module path, lexicographically.
// 2. then, if module paths are equal, then by their workspace name, lexicographically
func Sort(getter moduleGetter) func(*Workspace, *Workspace) int {
	return func(i, j *Workspace) int {
		if i.ModuleID() == j.ModuleID() {
			// same module, compare workspace name
			switch {
			case i.Name < j.Name:
				return -1
			case i.Name > j.Name:
				return 1
			default:
				// same workspace (unlikely)
				return 0
			}
		}
		imod, _ := getter.Get(i.ModuleID())
		jmod, _ := getter.Get(j.ModuleID())

		if imod.Path < jmod.Path {
			return -1
		} else {
			return 1
		}
	}
}
