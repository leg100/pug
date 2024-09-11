package workspace

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

type reloader struct {
	*Service
}

type ReloadSummary struct {
	Added   []string
	Removed []string
}

func (s ReloadSummary) String() string {
	return fmt.Sprintf("+%d-%d", len(s.Added), len(s.Removed))
}

func (s ReloadSummary) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("workspaces.added", s.Added),
		slog.Any("workspaces.removed", s.Removed),
	)
}

func (r *reloader) createReloadTask(moduleID resource.ID) error {
	spec, err := r.Reload(moduleID)
	if err != nil {
		return err
	}
	_, err = r.tasks.Create(spec)
	return err
}

// Reload returns a task spec that runs `terraform workspace list` on a
// module and updates pug with the results, adding any newly discovered
// workspaces and pruning any workspaces no longer found to exist.
//
// TODO: separate into Load and Reload
func (r *reloader) Reload(moduleID resource.ID) (task.Spec, error) {
	mod, err := r.modules.Get(moduleID)
	if err != nil {
		return task.Spec{}, err
	}
	return task.Spec{
		ModuleID: &mod.ID,
		Path:     mod.Path,
		Execution: task.Execution{
			TerraformCommand: []string{"workspace", "list"},
		},
		BeforeExited: func(t *task.Task) (task.Summary, error) {
			found, current, err := parseList(t.NewReader(false))
			if err != nil {
				return nil, err
			}
			added, removed, err := r.resetWorkspaces(mod, found, current)
			if err != nil {
				return nil, err
			}
			return ReloadSummary{Added: added, Removed: removed}, nil
		},
	}, nil
}

// resetWorkspaces resets the workspaces for a module, adding newly discovered
// workspaces, removing workspaces that no longer exist, and setting the current
// workspace for the module.
func (r *reloader) resetWorkspaces(mod *module.Module, discovered []string, current string) (added []string, removed []string, err error) {
	// Gather existing workspaces for the module.
	var existing []*Workspace
	for _, ws := range r.table.List() {
		if ws.ModuleID == mod.ID {
			existing = append(existing, ws)
		}
	}

	// Add discovered workspaces that don't exist in pug
	for _, name := range discovered {
		if !slices.ContainsFunc(existing, func(ws *Workspace) bool {
			return ws.Name == name
		}) {
			add, err := New(mod, name)
			if err != nil {
				return nil, nil, fmt.Errorf("adding workspace: %w", err)
			}
			r.table.Add(add.ID, add)
			added = append(added, name)
		}
	}
	// Remove workspaces from pug that no longer exist
	for _, ws := range existing {
		if !slices.Contains(discovered, ws.Name) {
			r.table.Delete(ws.ID)
			removed = append(removed, ws.Name)
		}
	}
	// Reset current workspace
	currentWorkspace, err := r.GetByName(mod.Path, current)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot find current workspace: %s: %w", current, err)
	}
	err = r.modules.SetCurrent(mod.ID, currentWorkspace.ID)
	if err != nil {
		return nil, nil, err
	}
	return
}

// Parse workspaces from the output of `terraform workspace list`.
//
// The output should contain something like this:
//
//	<asterisk> default
//	  non-default-1
//	  non-default-2
func parseList(r io.Reader) (list []string, current string, err error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if _, name, found := strings.Cut(scanner.Text(), "* "); found {
			// An asterisk prefix means this is the current workspace.
			current = name
			list = append(list, name)
		} else if _, name, found := strings.Cut(scanner.Text(), "  "); found {
			list = append(list, name)
		} else {
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, "", err
	}
	return
}
