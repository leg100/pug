package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWorkspace_SetCurrentWorkspace(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModuleWithTwoWorkspaces(t)

	// Navigate two children down in the tree - the cursor should be on the
	// module, and default - the current workspace - should be the next child -
	// and then the workspace we want to set as the new current workspace - dev
	// - is the last child.
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	// Make dev the current workspace
	tm.Type("C")

	// Expect dev to be the new current workspace
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "set current workspace to dev")
	})
}

func TestWorkspace_SinglePlan(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModule_Explorer(t)

	// Place cursor on workspace
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	// Create plan on default workspace
	tm.Type("p")

	// Expect to be taken to the plan's task page
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "plan 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +10~0-0")
	})
}

func TestWorkspace_MultiplePlans(t *testing.T) {
	t.Parallel()

	tm := setupAndInitMultipleModules(t)

	// Place cursor on module a's default workspace
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	// Create plan on all four workspaces
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("p")

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "plan 4/4") &&
			matchPattern(t, `modules/a.*default.*plan.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/a.*dev.*plan.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/b.*default.*plan.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/c.*default.*plan.*exited.*\+10~0-0`, s)
	})
}

func TestWorkspace_SingleApply(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModule_Explorer(t)

	// Place cursor on workspace
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	// Create apply on workspace
	tm.Type("a")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Auto-apply 1 workspaces? (y/N):")
	})
	tm.Type("y")

	// Send to apply task page
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "apply 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +10~0-0")
	})
}

func TestWorkspace_MultipleApplies(t *testing.T) {
	t.Parallel()

	tm := setupAndInitMultipleModules(t)

	// Place cursor on module a's default workspace
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	// Create apply on all four workspaces
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("a")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Auto-apply 4 workspaces? (y/N):")
	})
	tm.Type("y")

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "apply 4/4") &&
			matchPattern(t, `modules/a.*default.*apply.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/a.*dev.*apply.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/b.*default.*apply.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/c.*default.*apply.*exited.*\+10~0-0`, s)
	})
}

func TestWorkspace_SingleDestroy(t *testing.T) {
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

	// Go back to explorer and place cursor on default workspace
	tm.Type("0")
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
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

func TestWorkspace_MultipleDestroy(t *testing.T) {
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

	// Go back to explorer and place cursor on default workspace
	tm.Type("0")
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	// Destroy all resources on all three workspaces
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("D")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Destroy resources of 3 workspaces? (y/N):")
	})
	tm.Type("y")

	// Send to task group page
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "apply (destroy) 3/3") &&
			matchPattern(t, `modules/a.*default.*apply \(destroy\).*exited.*\+0~0-10`, s) &&
			matchPattern(t, `modules/b.*default.*apply \(destroy\).*exited.*\+0~0-10`, s) &&
			matchPattern(t, `modules/c.*default.*apply \(destroy\).*exited.*\+0~0-10`, s)
	})
}

func TestWorkspace_Delete(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModuleWithTwoWorkspaces(t)

	// Navigate two children down in the tree - the cursor should be on the
	// module, and default - the current workspace - should be the next child -
	// and then the workspace we want to set as the new current workspace - dev
	// - is the last child.
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	// Delete dev workspace
	tm.Send(tea.KeyMsg{Type: tea.KeyDelete})

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Delete workspace dev? (y/N):")
	})
	tm.Type("y")

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "workspace delete 󰠱 modules/a") &&
			strings.Contains(s, `Deleted workspace "dev"!`)
	})
}

func setupAndInitModuleWithTwoWorkspaces(t *testing.T) *testModel {
	tm := setup(t, "./testdata/module_with_two_workspaces")

	// Expect single module to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "└ 󰠱 a")
	})

	// Initialize module
	tm.Type("i")

	// Expect init to succeed, and to populate pug with two workspaces with 0
	// resources each
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!") &&
			strings.Contains(s, "init 󰠱 modules/a") &&
			strings.Contains(s, "exited") &&
			strings.Contains(s, "├  default 0") &&
			strings.Contains(s, "└  dev 0")
	})

	// Go back to explorer
	tm.Type("0")

	return tm
}
