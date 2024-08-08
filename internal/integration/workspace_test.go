package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWorkspace_SetCurrentWorkspace(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModuleWithTwoWorkspaces(t)

	// Navigate one row down to the dev workspace (default should be the first
	// row because it is lexicographically before dev).
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	// Make dev the current workspace
	tm.Type("C")

	// Expect dev to be the new current workspace
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `modules/a.*default.*`, s) &&
			matchPattern(t, `modules/a.*dev.*✓`, s)
	})
}

func TestWorkspace_SinglePlan(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModule(t)

	// Go to workspace listing
	tm.Type("w")

	// Wait for modules/a's default workspace to be listed. This should be the
	// first workspace listed.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Workspaces") &&
			matchPattern(t, `modules/a.*default.*✓`, s)
	})

	// Create plan on modules/a's default workspace
	tm.Type("p")

	// Expect to be taken to the plan's task page
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `Task.*plan.*default.*modules/a.*\+10~0-0.*exited`, s)
	})
}

func TestWorkspace_MultiplePlans(t *testing.T) {
	t.Parallel()

	// Initialize all modules
	tm := setupAndInitModuleList(t)

	// Go to workspace listing
	tm.Type("w")

	// Wait for all four workspaces to be listed.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `modules/a.*default.*`, s) &&
			matchPattern(t, `modules/a.*dev`, s) &&
			matchPattern(t, `modules/b.*default`, s) &&
			matchPattern(t, `modules/c.*default`, s)
	})

	// Create plan on all four workspaces
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("p")

	// Expect to be taken to task group's page
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*plan.*4/4", s) &&
			matchPattern(t, `modules/a.*default.*\+10~0-0`, s) &&
			matchPattern(t, `modules/a.*dev.*\+10~0-0`, s) &&
			matchPattern(t, `modules/b.*default.*\+10~0-0`, s) &&
			matchPattern(t, `modules/c.*default.*\+10~0-0`, s)
	})

}

func TestWorkspace_SingleApply(t *testing.T) {
	t.Parallel()

	// Initialize module
	tm := setupAndInitModule(t)

	// Go to workspace listing
	tm.Type("w")

	// Wait for modules/a's default workspace to be listed. This should be the
	// first workspace listed.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Workspaces") &&
			matchPattern(t, `modules/a.*default.*✓`, s)
	})

	// Create apply on workspace
	tm.Type("a")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Auto-apply 1 workspaces? (y/N):")
	})
	tm.Type("y")

	// Send to apply task page
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `Task.*apply.*default.*modules/a.*\+10~0-0.*exited`, s)
	})

}

func TestWorkspace_MultipleApplies(t *testing.T) {
	t.Parallel()

	// Initialize all modules
	tm := setupAndInitModuleList(t)

	// Go to workspace listing
	tm.Type("w")

	// Wait for all four workspaces to be listed.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `modules/a.*default.*`, s) &&
			matchPattern(t, `modules/a.*dev`, s) &&
			matchPattern(t, `modules/b.*default`, s) &&
			matchPattern(t, `modules/c.*default`, s)
	})

	// Select all workspaces and create apply on each
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("a")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Auto-apply 4 workspaces? (y/N):")
	})
	tm.Type("y")

	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*apply.*4/4", s) &&
			matchPattern(t, `modules/a.*default.*\+10~0-0`, s) &&
			matchPattern(t, `modules/a.*dev.*\+10~0-0`, s) &&
			matchPattern(t, `modules/b.*default.*\+10~0-0`, s) &&
			matchPattern(t, `modules/c.*default.*\+10~0-0`, s)
	})
}

func TestWorkspace_SingleDestroy(t *testing.T) {
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

	// Go to workspace listing
	tm.Type("w")

	// Workspace should have 10 resources in its state loaded.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Workspaces") &&
			matchPattern(t, `modules/a.*default.*10`, s)
	})

	// Destroy all resources in workspace
	tm.Type("d")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Destroy resources of 1 workspaces? (y/N):")
	})
	tm.Type("y")

	// Send to apply task page
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `Task.*apply.*default.*modules/a.*\+0~0-10.*exited`, s)
	})
}

func TestWorkspace_MultipleDestroy(t *testing.T) {
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

	// Go to workspace listing
	tm.Type("w")

	// Expect three workspaces to be listed, each with 10 resources
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Workspaces") &&
			matchPattern(t, `modules/a.*default.*10`, s) &&
			matchPattern(t, `modules/b.*default.*10`, s) &&
			matchPattern(t, `modules/c.*default.*10`, s)
	})

	// Destroy all resources in all three workspaces
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("d")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Destroy resources of 3 workspaces? (y/N):")
	})
	tm.Type("y")

	// Send to task group page
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*apply.*3/3", s) &&
			matchPattern(t, `modules/a.*default.*\+0~0-10`, s) &&
			matchPattern(t, `modules/b.*default.*\+0~0-10`, s) &&
			matchPattern(t, `modules/c.*default.*\+0~0-10`, s)
	})
}

func TestWorkspace_Filter(t *testing.T) {
	t.Parallel()

	// Initialize all modules
	tm := setupAndInitModuleList(t)

	// Go to workspaces listing
	tm.Type("w")

	// Wait for all four workspaces to be listed.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `modules/a.*default`, s) &&
			matchPattern(t, `modules/a.*dev`, s) &&
			matchPattern(t, `modules/b.*default`, s) &&
			matchPattern(t, `modules/c.*default`, s)
	})

	// Focus filter widget
	tm.Type("/")

	// Expect filter prompt
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Filter:")
	})

	// Filter to only show modules/a
	tm.Type("modules/a")

	// Expect title to show 2 workspaces filtered out of a total 4.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "2/4")
	})
}

func TestWorkspace_Delete(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModuleWithTwoWorkspaces(t)

	// Filter workspaces to only show dev workspace. This is the only way to
	// ensure the dev workspace is currently highlighted.

	// Focus filter widget
	tm.Type("/")

	// Expect filter prompt
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Filter:")
	})

	// Filter to only show dev workspace
	tm.Type("dev")

	// Expect title to show 1 workspaces filtered out of a total of 2.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "1/2")
	})

	// Exit filter
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Now the dev workspace should be highlighted.

	// Delete dev workspace
	tm.Type("D")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Delete 1 workspace(s)? (y/N):")
	})
	tm.Type("y")

	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `Task.*workspace delete.*modules/a.*exited`, s) &&
			strings.Contains(s, `Deleted workspace "dev"!`)
	})
}

func setupAndInitModuleWithTwoWorkspaces(t *testing.T) *testModel {
	tm := setup(t, "./testdata/module_with_two_workspaces")

	// Wait for module to be loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!")
	})

	// Go to workspace listing
	tm.Type("w")

	// Expect two workspaces to be listed, and expect default to be the current
	// workspace
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `modules/a.*default.*✓`, s) &&
			matchPattern(t, `modules/a.*dev`, s)
	})

	return tm
}
