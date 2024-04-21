package app

import (
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
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

	// Expect to see summary of changes
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "+10~0-0")
	})

	// Apply plan
	tm.Type("a")

	// Wait for apply to complete
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply complete! Resources: 10 added, 0 changed, 0 destroyed.")
	})
}
