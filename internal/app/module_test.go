package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModule_Loaded(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Expect message to inform user that three modules have been loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "loaded 3 modules")
	})
}

func TestModule_SingleInit(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Expect module to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")

	// Expect to see successful init message
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!")
	})
}

func TestModule_MultipleInit(t *testing.T) {
	t.Parallel()

	setupAndInitModuleList(t)
}

func TestModule_MultipleFormat(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Expect three modules to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a") &&
			strings.Contains(s, "modules/b") &&
			strings.Contains(s, "modules/c")
	})

	// Format all modules
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("f")
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*format", s) &&
			matchPattern(t, `modules/a.*exited`, s) &&
			matchPattern(t, `modules/b.*exited`, s) &&
			matchPattern(t, `modules/c.*exited`, s)
	})
}

func TestModule_MultipleValidate(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModuleList(t)

	// Validate all modules
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("v")
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*validate", s) &&
			matchPattern(t, `modules/a.*exited`, s) &&
			matchPattern(t, `modules/b.*exited`, s) &&
			matchPattern(t, `modules/c.*exited`, s)
	})
}

func TestModule_Reload(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Expect message to inform user that three modules have been loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "loaded 3 modules")
	})

	// Reload modules
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Expect message to inform user that reload has finished and no modules
	// have been added nor removed.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "reloaded modules: added 0; removed 0")
	})
}

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
		return matchPattern(t, "TaskGroup.*reload-workspace", s) &&
			matchPattern(t, "modules/a.*exited", s) &&
			matchPattern(t, "modules/b.*exited", s) &&
			matchPattern(t, "modules/c.*exited", s)
	})
}

// TODO: test creating plan on an uninitialized module(s)

func TestModule_SinglePlan(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModule(t)

	// Create plan on first module
	tm.Type("p")
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "Task.*plan.*default.*modules/a.*exited", s)
	})

}

func TestModule_MultiplePlans(t *testing.T) {
	t.Parallel()

	// Initialize all modules
	tm := setupAndInitModuleList(t)

	// Select all modules and invoke plan.
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("p")
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*plan.*3/3", s) &&
			matchPattern(t, `modules/a.*default.*\+10~0-0`, s) &&
			matchPattern(t, `modules/b.*default.*\+10~0-0`, s) &&
			matchPattern(t, `modules/c.*default.*\+10~0-0`, s)
	})

}

func TestModule_SingleDestroyPlan(t *testing.T) {
	t.Parallel()

	// Setup test with pre-existing state
	tm := setup(t, "./testdata/module_destroy")

	// Wait for module to be loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")

	// Expect user to be taken to init's task page.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!")
	})

	// Go back to module listing
	tm.Type("m")

	// Module should have 10 resources in its state loaded.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `modules/a.*default.*10`, s)
	})

	// Create destroy plan
	tm.Type("d")

	// Expect 10 resources to be proposed for deletion
	waitFor(t, tm, func(s string) bool {
		//s = internal.StripAnsi(s)
		return matchPattern(t, `Task.*plan.*default.*modules/a.*\+0~0\-10.*exited`, s)
	})
}

func TestModule_SingleApply(t *testing.T) {
	t.Parallel()

	// Initialize single module
	tm := setupAndInitModule(t)

	// Create apply on whatever the currently highlighted module is
	tm.Type("a")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Auto-apply 1 modules? (y/N):")
	})
	tm.Type("y")

	// Send to apply task page
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `Task.*apply.*default.*modules/a.*\+10~0-0.*exited`, s)
	})

}

func TestModule_MultipleApplies(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModuleList(t)

	// Select all modules and create apply on each
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("a")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Auto-apply 3 modules? (y/N):")
	})
	tm.Type("y")

	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*apply.*3/3", s) &&
			matchPattern(t, `modules/a.*default.*\+10~0-0`, s) &&
			matchPattern(t, `modules/b.*default.*\+10~0-0`, s) &&
			matchPattern(t, `modules/c.*default.*\+10~0-0`, s)
	})
}

// TODO: test pruning: attempt to create plans / applies on multiple modules,
// some of which are uninitialized. User should receive message that their
// selection has been pruned.

func TestModuleList_Filter(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModuleList(t)

	// Focus filter widget
	tm.Type("/")

	// Expect filter prompt
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Filter:")
	})

	// Filter to only show modules/a
	tm.Type("modules/a")

	// Expect table title to show 1 module filtered out of a total 3.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "1/3")
	})
}

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
