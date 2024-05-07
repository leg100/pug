package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/leg100/pug/internal"
)

func TestWorkspace_Resources(t *testing.T) {
	t.Parallel()

	tm := setupResourcesTab(t)

	// Expect to see several resources listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "random_pet.pet[0]") &&
			strings.Contains(s, "random_pet.pet[1]") &&
			strings.Contains(s, "random_pet.pet[2]")
	})

	// Taint several resources
	tm.Type("t")
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Type("t")
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Type("t")

	// Expect to see several resources tainted
	waitFor(t, tm, func(s string) bool {
		return strings.Count(s, "tainted") == 3
	})

	// Untaint several resources
	tm.Type("u")
	tm.Send(tea.KeyMsg{Type: tea.KeyUp})
	tm.Type("u")
	tm.Send(tea.KeyMsg{Type: tea.KeyUp})
	tm.Type("u")

	// Expect to see no resources tainted
	waitFor(t, tm, func(s string) bool {
		return strings.Count(s, "tainted") == 0
	})

	// Select several resources and delete them
	tm.Type("s")
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Type("s")
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Type("s")
	tm.Type("D")

	// Confirm deletion
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Delete 3 resource(s) (y/N)?")
	})
	tm.Type("y")

	// Expect only 7 resources. Note we can't test that the three deleted
	// resources are NOT listed because waitFor accumulates all the string
	// output since it was called, which is likely to include resources from
	// both before and after the deletion. So instead we check for the presence
	// of a new total number of resources.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "resources (7)")
	})

	// Select several resources and create targeted plan
	tm.Type("s")
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Type("s")
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Type("s")
	tm.Type("p")

	// Expect to be taken to the run page, with a completed plan, and a warning
	// that resource targeting is in effect
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return strings.Contains(s, "plan ✓") &&
			strings.Contains(s, "Warning: Resource targeting is in effect")
	})
}

func TestWorkspace_Resources_Filter(t *testing.T) {
	t.Parallel()

	tm := setupResourcesTab(t)

	// Expect to see several resources listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "random_pet.pet[0]") &&
			strings.Contains(s, "random_pet.pet[1]") &&
			strings.Contains(s, "random_pet.pet[2]")
	})

	// Focus filter widget
	tm.Type("/")

	// Expect filter prompt
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Filter:")
	})

	// Filter to only show pet[1]
	tm.Type("pet[1]")

	// Expect resources tab to show 1 resources filtered out of a total 10.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "resources (1/10)")
	})
}

func setupResourcesTab(t *testing.T) *teatest.TestModel {
	// Setup test with pre-existing state
	tm := setup(t, "./testdata/workspace_resources")

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
		return matchPattern(t, `modules/a.*default`, s)
	})

	// Go to workspace's page
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Expect resources tab title
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "resources (10)")
	})

	// Select resources tab (two tabs to the right).
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})

	return tm
}
