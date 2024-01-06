package app

import (
	"strings"
	"testing"

	"github.com/leg100/pug/internal"
)

// TestRunList_Single tests interacting with a single run in the run list view.
func TestRunList_Single(t *testing.T) {
	tm := setup(t)

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
		t.Log(s)
		return strings.Contains(s, "Run[default](modules/a)")
		// TODO: expect 'apply <tick>' tab header
	})

	// Wait for apply to complete
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply complete! Resources: 10 added, 0 changed, 0 destroyed.")
	})
}
