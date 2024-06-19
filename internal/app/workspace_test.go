package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWorkspace_SetCurrentWorkspace(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/set_current_workspace")

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

	tm := setup(t, "./testdata/module_list")

	// Initialize all modules
	initAllModules(t, tm)

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
		return matchPattern(t, "Task.*plan.*default.*modules/a.*exited.*planned", s)
	})
}

func TestWorkspace_MultiplePlans(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Initialize all modules
	initAllModules(t, tm)

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
			matchPattern(t, "modules/a.*default.*planned", s) &&
			matchPattern(t, "modules/a.*dev.*planned", s) &&
			matchPattern(t, "modules/b.*default.*planned", s) &&
			matchPattern(t, "modules/c.*default.*planned", s)
	})

}

func TestWorkspace_SingleApply(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Initialize all modules
	initAllModules(t, tm)

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
		return matchPattern(t, "Task.*apply.*default.*modules/a.*exited.*applied", s)
	})

}

func TestWorkspace_MultipleApplies(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Initialize all modules
	initAllModules(t, tm)

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
			matchPattern(t, "modules/a.*default.*applied", s) &&
			matchPattern(t, "modules/a.*dev.*applied", s) &&
			matchPattern(t, "modules/b.*default.*applied", s) &&
			matchPattern(t, "modules/c.*default.*applied", s)
	})
}

func TestWorkspace_Filter(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Initialize all modules
	initAllModules(t, tm)

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
