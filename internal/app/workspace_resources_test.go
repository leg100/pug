package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal"
)

func TestWorkspace_Resources(t *testing.T) {
	tm := setup(t)

	// Initialize and apply run on modules/a
	initAndApplyModuleA(t, tm)

	// Go to workspaces
	tm.Type("W")

	// Wait for workspace to be loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "default")
	})

	// Select the default workspace (should be the first and only workspace
	// listed)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Expect resources tab title
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "resources (10)")
	})

	// Select resources tab (two tabs to the right).
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})

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
	tm.Type("d")

	// Confirm deletion
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Delete 3 resource(s) (y/N)?")
	})
	tm.Type("y")

	// Expect to no longer see first three pets listed
	waitFor(t, tm, func(s string) bool {
		return !strings.Contains(s, "random_pet.pet[0]") &&
			!strings.Contains(s, "random_pet.pet[1]") &&
			!strings.Contains(s, "random_pet.pet[2]")
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
