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

	tm := setupAndInitModule_Explorer(t)

	// Reload workspaces for module
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlW})

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "workspace list 󰠱 modules/a") &&
			strings.Contains(s, "* default")
	})
}

func TestModuleList_ReloadWorkspacesMultipleModules(t *testing.T) {
	t.Parallel()

	tm := setupAndInitMultipleModules(t)

	// Select all modules and reload workspaces for each and every module
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlW})

	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "workspace list 3/3", s) &&
			matchPattern(t, "modules/a.*workspace list.*exited", s) &&
			matchPattern(t, "modules/b.*workspace list.*exited", s) &&
			matchPattern(t, "modules/c.*workspace list.*exited", s)
	})
}

func TestExplorer_SingleInit(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/single_module")

	// Expect single module in tree
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "└ 󰠱 a")
	})

	// Cursor should automatically be on module.
	// Initialize module.
	tm.Type("i")

	// Expect to see successful init message.
	// Tree should now have workspace under module.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!") &&
			strings.Contains(s, "init 󰠱 modules/a") &&
			strings.Contains(s, "└ 󰠱 a") &&
			strings.Contains(s, "└  default")
	})
}

func TestExplorer_MultipleInit(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Expect three modules in tree
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "├ 󰠱 a") &&
			strings.Contains(s, "├ 󰠱 b") &&
			strings.Contains(s, "└ 󰠱 c")
	})

	// Select all modules and init
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("i")

	// Expect init task group with 3 successful tasks
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "init 3/3") &&
			matchPattern(t, `modules/a.*init.*exited`, s) &&
			matchPattern(t, `modules/b.*init.*exited`, s) &&
			matchPattern(t, `modules/c.*init.*exited`, s)
	})
}

func TestExplorer_SingleInitUpgrade(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/single_module")

	// Expect single module in tree
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "└ 󰠱 a")
	})

	// Cursor should automatically be on module.
	// Initialize module, upgrading any providers
	tm.Type("u")

	// Expect to see task header
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "init 󰠱 modules/a")
	})

	// Show task info sidebar
	tm.Type("I")

	// Expect to see -upgrade argument
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "-upgrade")
	})
}

func TestExplorer_MultipleFormat(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Expect three modules in tree
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "├ 󰠱 a") &&
			strings.Contains(s, "├ 󰠱 b") &&
			strings.Contains(s, "└ 󰠱 c")
	})

	// Format all modules
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("f")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "fmt 3/3") &&
			matchPattern(t, `modules/a.*fmt.*exited`, s) &&
			matchPattern(t, `modules/b.*fmt.*exited`, s) &&
			matchPattern(t, `modules/c.*fmt.*exited`, s)
	})
}

func TestExplorer_MultipleValidate(t *testing.T) {
	t.Parallel()

	tm := setupAndInitMultipleModules(t)

	// Validate all modules
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("v")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "validate 3/3") &&
			matchPattern(t, `modules/a.*validate.*exited`, s) &&
			matchPattern(t, `modules/b.*validate.*exited`, s) &&
			matchPattern(t, `modules/c.*validate.*exited`, s)
	})
}

func TestExplorer_SinglePlan(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModule_Explorer(t)

	// Create plan on module
	tm.Type("p")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "plan 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +10~0-0")
	})
}

func TestExplorer_MultiplePlans(t *testing.T) {
	t.Parallel()

	tm := setupAndInitMultipleModules(t)

	// Select all modules and invoke plan.
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("p")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "plan 3/3") &&
			matchPattern(t, `modules/a.*default.*plan.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/b.*default.*plan.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/c.*default.*plan.*exited.*\+10~0-0`, s)
	})
}

func TestExplorer_SingleApply(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModule_Explorer(t)

	// Create apply
	tm.Type("a")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		t.Log(s)
		return strings.Contains(s, "Auto-apply 1 workspaces? (y/N):")
	})
	tm.Type("y")

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "apply 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +10~0-0")
	})
}

func TestExplorer_MultipleApplies(t *testing.T) {
	t.Parallel()

	tm := setupAndInitMultipleModules(t)

	// Select all modules and invoke applies.
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("a")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Auto-apply 3 workspaces? (y/N):")
	})
	tm.Type("y")

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "apply 3/3") &&
			matchPattern(t, `modules/a.*default.*apply.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/b.*default.*apply.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/c.*default.*apply.*exited.*\+10~0-0`, s)
	})
}

func TestExplorer_SingleDestroyPlan(t *testing.T) {
	t.Parallel()

	// Setup test with pre-existing state
	tm := setup(t, "./testdata/module_destroy")

	// Expect single module to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "└ 󰠱 a")
	})

	// Initialize module
	tm.Type("i")

	// Init should finish successfully and there should now be a workspace
	// listed in the tree with 10 resources.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!") &&
			strings.Contains(s, "init 󰠱 modules/a") &&
			strings.Contains(s, "exited") &&
			strings.Contains(s, "└ 󰠱 a") &&
			strings.Contains(s, "└  default 10")
	})

	// Create destroy plan
	tm.Type("0")
	tm.Type("d")

	// Expect 10 resources to be proposed for deletion
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "plan (destroy) 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +0~0-10")
	})
}

func TestExplorer_SingleDestroy(t *testing.T) {
	t.Parallel()

	// Setup test with pre-existing state
	tm := setup(t, "./testdata/module_destroy")

	// Expect single module to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "└ 󰠱 a")
	})

	// Initialize module
	tm.Type("i")

	// Init should finish successfully and there should now be a workspace
	// listed in the tree with 10 resources.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!") &&
			strings.Contains(s, "init 󰠱 modules/a") &&
			strings.Contains(s, "exited") &&
			strings.Contains(s, "└ 󰠱 a") &&
			strings.Contains(s, "└  default 10")
	})

	// Create destroy
	tm.Type("0")
	tm.Type("D")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Destroy resources of 1 workspaces? (y/N):")
	})
	tm.Type("y")

	// Expect destroy task to result in destruction of 10 resources
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "apply (destroy) 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +0~0-10") &&
			strings.Contains(s, "└  default 0")
	})
}

func TestExplorer_MultipleDestroys(t *testing.T) {
	t.Parallel()

	// Setup test with modules with pre-existing state
	tm := setup(t, "./testdata/multiple_destroy")

	// Expect three modules to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "├ 󰠱 a") &&
			strings.Contains(s, "├ 󰠱 b") &&
			strings.Contains(s, "└ 󰠱 c")
	})

	// Select all modules and init
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("i")

	// Each module should now be populated with at least one workspace.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "󰠱 3  3")
	})

	// Go back to explorer
	tm.Type("0")

	// Select all modules and invoke destroy.
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("D")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Destroy resources of 3 workspaces? (y/N):")
	})
	tm.Type("y")

	waitFor(t, tm, func(s string) bool {
		t.Log(s)
		return strings.Contains(s, "apply (destroy) 3/3") &&
			matchPattern(t, `modules/a.*default.*apply \(destroy\).*exited.*\+0~0-10`, s) &&
			matchPattern(t, `modules/b.*default.*apply \(destroy\).*exited.*\+0~0-10`, s) &&
			matchPattern(t, `modules/c.*default.*apply \(destroy\).*exited.*\+0~0-10`, s)
	})
}

func TestExplorer_WithVars(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_with_vars")

	// Expect single module to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "└ 󰠱 a")
	})

	// Initialize module
	tm.Type("i")

	// Init should finish successfully and there should now be a workspace
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!") &&
			strings.Contains(s, "init 󰠱 modules/a") &&
			strings.Contains(s, "exited") &&
			strings.Contains(s, "└ 󰠱 a") &&
			strings.Contains(s, "└  default 0")
	})

	// Create plan for default workspace
	tm.Type("0")
	tm.Type("p")

	// Expect to see summary of changes
	waitFor(t, tm, func(s string) bool {
		// Remove formatting
		s = internal.StripAnsi(s)
		return strings.Contains(s, "plan 󰠱 modules/a  default") &&
			strings.Contains(s, "Changes to Outputs:") &&
			strings.Contains(s, `+ foo = "override"`) &&
			strings.Contains(s, "exited +0~0-0")
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

func TestExplorer_MultipleExecute(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Expect three modules in tree
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "├ 󰠱 a") &&
			strings.Contains(s, "├ 󰠱 b") &&
			strings.Contains(s, "└ 󰠱 c")
	})

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
		return strings.Contains(s, "terraform 3/3") &&
			matchPattern(t, `modules/a.*terraform.*exited`, s) &&
			matchPattern(t, `modules/b.*terraform.*exited`, s) &&
			matchPattern(t, `modules/c.*terraform.*exited`, s)
	})
}

func setupAndInitModule_Explorer(t *testing.T) *testModel {
	tm := setup(t, "./testdata/single_module")

	// Expect single module to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "└ 󰠱 a")
	})

	// Initialize module
	tm.Type("i")
	// Expect init to succeed, and to populate pug with one workspace with 0
	// resources
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!") &&
			strings.Contains(s, "init 󰠱 modules/a") &&
			strings.Contains(s, "exited") &&
			strings.Contains(s, "└  default 0")
	})

	// Go back to explorer
	tm.Type("0")

	return tm
}

func setupAndInitMultipleModules(t *testing.T) *testModel {
	// TODO: helper?
	//
	tm := setup(t, "./testdata/module_list")

	// Expect three modules in tree
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "├ 󰠱 a") &&
			strings.Contains(s, "├ 󰠱 b") &&
			strings.Contains(s, "└ 󰠱 c")
	})

	// Select all modules and init
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("i")

	// Expect init task group with 3 successful tasks.
	// Each module should now be populated with at least one workspace.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "init 3/3") &&
			matchPattern(t, `modules/a.*init.*exited`, s) &&
			matchPattern(t, `modules/b.*init.*exited`, s) &&
			matchPattern(t, `modules/c.*init.*exited`, s) &&
			strings.Contains(s, "󰠱 3  4")
	})

	// Go back to explorer
	tm.Type("0")

	// Clear selection
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlBackslash})

	return tm
}
