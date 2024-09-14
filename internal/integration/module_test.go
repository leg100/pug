package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal"
)

func TestModule_Loaded(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Expect message to inform user that three modules have been loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "loaded 3 modules")
	})
}

func TestModule_Backend(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Expect three modules, each with a local backend
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `modules/a.*local`, s) &&
			matchPattern(t, `modules/b.*local`, s) &&
			matchPattern(t, `modules/c.*local`, s)
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

func TestModule_SingleInitUpgrade(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Expect module to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module, upgrading any providers
	tm.Type("u")

	// Expect to see successful init message
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!")
	})

	// Show task info sidebar
	tm.Type("I")

	// Expect to see -upgrade argument
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "-upgrade")
	})
}

func TestModule_MultipleInitUpgrade(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Go to modules listing
	tm.Type("m")

	// Expect three modules to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a") &&
			strings.Contains(s, "modules/b") &&
			strings.Contains(s, "modules/c")
	})

	// Select all modules and run `init -upgrade`
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("u")
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*init", s) &&
			matchPattern(t, `modules/a.*exited`, s) &&
			matchPattern(t, `modules/b.*exited`, s) &&
			matchPattern(t, `modules/c.*exited`, s)
	})

	// Show task info sidebar
	tm.Type("I")

	// Expect to see -upgrade argument
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "-upgrade")
	})
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
		return matchPattern(t, "TaskGroup.*fmt", s) &&
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
		return matchPattern(t, "TaskGroup.*workspace list", s) &&
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
	tm.Type("P")

	// Expect 10 resources to be proposed for deletion
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `Task.*plan \(destroy\).*default.*modules/a.*\+0~0\-10.*exited`, s)
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

func TestModule_SingleDestroy(t *testing.T) {
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

	// Create destroy
	tm.Type("d")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Destroy resources of 1 modules? (y/N):")
	})
	tm.Type("y")

	// Send to apply task page
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `Task.*apply \(destroy\).*default.*modules/a.*\+0~0-10.*exited`, s)
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

func TestModule_MultipleDestroy(t *testing.T) {
	t.Parallel()

	// Setup test with modules with pre-existing state
	tm := setup(t, "./testdata/multiple_destroy")

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

	// Destroy resources of all 3 modules (all modules should still be selected)
	tm.Type("d")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Destroy resources of 3 modules? (y/N):")
	})
	tm.Type("y")

	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `TaskGroup.*apply \(destroy\).*3/3`, s) &&
			matchPattern(t, `modules/a.*default.*\+0~0-10`, s) &&
			matchPattern(t, `modules/b.*default.*\+0~0-10`, s) &&
			matchPattern(t, `modules/c.*default.*\+0~0-10`, s)
	})
}

func TestModule_WithVars(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_with_vars")

	// Wait for module to be loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!")
	})

	// Go to module page
	tm.Type("m")

	// Wait for default workspace to be loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "default")
	})

	// Create plan for default workspace
	tm.Type("p")

	// Expect to see summary of changes
	waitFor(t, tm, func(s string) bool {
		// Remove formatting
		s = internal.StripAnsi(s)
		return matchPattern(t, `Task.*plan.*default.*modules/a.*\+0~0-0.*exited`, s) &&
			strings.Contains(s, "Changes to Outputs:") &&
			strings.Contains(s, `+ foo = "override"`)
	})

	// Apply plan and provide confirmation
	tm.Type("a")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply plan? (y/N):")
	})
	tm.Type("y")

	// Wait for apply to complete
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply complete! Resources: 0 added, 0 changed, 0 destroyed.") &&
			strings.Contains(s, `foo = "override"`)
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

func TestModule_MultipleExecute(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModuleList(t)

	// Select all modules and run program in each module directory.
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("x")

	// Expect prompt for program to run.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Execute program in 3 module directories: ")
	})
	tm.Type("terraform version\n")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*terraform.*3/3", s) &&
			matchPattern(t, `modules/a.*terraform`, s) &&
			matchPattern(t, `modules/b.*terraform`, s) &&
			matchPattern(t, `modules/c.*terraform`, s)
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
