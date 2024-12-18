package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModuleList_ReloadWorkspacesSingleModule(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModuleList(t)

	// Reload workspaces for what ever module is currently selected.
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlW})

	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "Task.*workspace list", s) &&
			strings.Contains(s, "* default")
	})
}

func TestModuleList_ReloadWorkspacesMultipleModules(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModuleList(t)

	// Select all modules and reload workspaces for each and every module
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlW})

	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*workspace list", s) &&
			matchPattern(t, "modules/a.*exited", s) &&
			matchPattern(t, "modules/b.*exited", s) &&
			matchPattern(t, "modules/c.*exited", s)
	})
}

// TODO: test creating plan on an uninitialized module(s)

// TODO: test pruning: attempt to create plans / applies on multiple modules,
// some of which are uninitialized. User should receive message that their
// selection has been pruned.

func setupAndInitModuleList(t *testing.T) *testModel {
	tm := setup(t, "./testdata/module_list")

	// Go to modules listing
	tm.Type("m")

	// Expect three modules to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a") &&
			strings.Contains(s, "modules/b") &&
			strings.Contains(s, "modules/c")
	})

	// Select all modules and init
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*init", s) &&
			matchPattern(t, `modules/a.*exited`, s) &&
			matchPattern(t, `modules/b.*exited`, s) &&
			matchPattern(t, `modules/c.*exited`, s)
	})

	// Go back to modules listing
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Expect three modules to be listed, along with their default workspace.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "modules/a.*default", s) &&
			matchPattern(t, "modules/b.*default", s) &&
			matchPattern(t, "modules/c.*default", s)
	})

	// Clear selection
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlBackslash})

	return tm
}

func setupAndInitModule(t *testing.T) *testModel {
	tm := setup(t, "./testdata/single_module")

	// Go to modules listing
	tm.Type("m")

	// Expect single module to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "Task.*init.*modules/a.*exited", s)
	})

	// Go back to modules listing
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Expect single module to be listed, along with its default workspace.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "modules/a.*default", s)
	})

	return tm
}
