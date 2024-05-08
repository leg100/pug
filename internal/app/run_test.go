package app

import (
	"strings"
	"testing"

	"github.com/leg100/pug/internal"
)

func TestRun(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Initialize and apply run on modules/a
	initAndApplyModuleA(t, tm)
}

// TestRun_Stale tests that a planned run is placed into the 'stale' state when
// a succeeding run is created.
func TestRun_Stale(t *testing.T) {
	t.Parallel()

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

	// Create plan for first workspace
	tm.Type("p")

	// User should now be taken to the run page...

	// Expect to see summary of changes
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return strings.Contains(s, "Plan: 10 to add, 0 to change, 0 to destroy.")
	})

	// Go to its workspace page
	tm.Type("w")

	// Expect one run in the planned state
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "runs (1)") && matchPattern(t, `planned.*\+10~0\-0`, s)
	})

	// Start another run
	tm.Type("p")

	// Expect to see summary of changes, again.
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return strings.Contains(s, "Plan: 10 to add, 0 to change, 0 to destroy.")
	})

	// Go to its workspace page
	tm.Type("w")

	// Expect two runs, one in the planned state, one in the stale state
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "runs (2)") &&
			matchPattern(t, `planned.*\+10~0\-0`, s) &&
			matchPattern(t, `stale.*\+10~0\-0`, s)
	})
}

func TestRun_WithVars(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/run_with_vars")

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

	// User should now be taken to the run page...

	// Expect to see summary of changes, and the run should be in the planned
	// state
	waitFor(t, tm, func(s string) bool {
		// Remove formatting
		s = internal.StripAnsi(s)
		return strings.Contains(s, "Changes to Outputs:") &&
			strings.Contains(s, `+ foo = "override"`) &&
			strings.Contains(s, "planned")
	})

	// Apply plan and provide confirmation
	tm.Type("a")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Proceed with apply? (y/N):")
	})
	tm.Type("y")

	// Wait for apply to complete
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply complete! Resources: 0 added, 0 changed, 0 destroyed.") &&
			strings.Contains(s, `foo = "override"`)
	})
}
