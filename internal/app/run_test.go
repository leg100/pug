package app

import (
	"strings"
	"testing"

	"github.com/leg100/pug/internal"
)

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

	// Expect to see summary of changes
	waitFor(t, tm, func(s string) bool {
		// Remove formatting
		s = internal.StripAnsi(s)
		return strings.Contains(s, "Changes to Outputs:") &&
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
