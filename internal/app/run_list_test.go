package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal"
)

// TestRunList_Single tests interacting with a single run in the run list view.
func TestRunList_Single(t *testing.T) {
	tm := setup(t, "./testdata/module_list")

	// Wait for module to be loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!")
	})

	// Go to workspaces
	tm.Type("W")

	// Wait for workspace to be loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "default")
	})

	// Create plan for first module
	tm.Type("p")

	// User should now be taken to the run page...

	// Expect to see summary of changes
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return strings.Contains(s, "Plan: 10 to add, 0 to change, 0 to destroy.")
	})

	// Go to the run list page
	tm.Type("R")

	// Wait for run list page to be populated with planned run
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "planned")
	})

	// Apply run and provide confirmation
	tm.Type("a")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Proceed with apply (y/N)?")
	})
	tm.Type("y")

	// Expect to be taken back to the run page
	waitFor(t, tm, func(s string) bool {
		// Expect run page breadcrumbs
		return strings.Contains(s, "Run[default](modules/a)")
		// TODO: expect 'apply <tick>' tab header
	})

	// Wait for apply to complete
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply complete! Resources: 10 added, 0 changed, 0 destroyed.")
	})
}

// TestRunList_Multiple demonstrates interacting with multiple runs on the run
// list page.
func TestRunList_Multiple(t *testing.T) {
	tm := setup(t, "./testdata/module_list")

	// Expect message to inform user that modules have been loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "loaded 3 modules")
	})

	// Select all modules and init
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("i")

	// Wait for each module to be initialized, and to have its current workspace
	// set (should be "default")
	waitFor(t, tm, func(s string) bool {
		return strings.Count(s, "default") == 3
	})

	// Go to global workspaces page
	tm.Type("W")

	// Wait for all four workspaces to be listed
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `(?s)WORKSPACE.*default.*dev.*default.*default`, s)
	})

	// Create run on all four workspaces
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("p")

	// Go to global runs page
	tm.Type("R")

	// Expect all four runs to enter the planned state.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `modules/a.*default.*planned`, s) &&
			matchPattern(t, `modules/a.*dev.*planned`, s) &&
			matchPattern(t, `modules/b.*default.*planned`, s) &&
			matchPattern(t, `modules/c.*default.*planned`, s)
	})

	// Apply all four runs
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("a")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply 4 runs (y/N)?")
	})
	tm.Type("y")

	// Expect all four runs to enter the applied state.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `modules/a.*default.*applied`, s) &&
			matchPattern(t, `modules/a.*dev.*applied`, s) &&
			matchPattern(t, `modules/b.*default.*applied`, s) &&
			matchPattern(t, `modules/c.*default.*applied`, s)
	})

	// Attempt to apply already-applied run.
	tm.Type("a")

	// Expect error because run is not applyable
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "run is not in the planned state")
	})
}
