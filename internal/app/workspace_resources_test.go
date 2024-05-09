package app

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_Resources_Taint(t *testing.T) {
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
}

func TestWorkspace_Resources_Move(t *testing.T) {
	t.Parallel()

	tm := setupResourcesTab(t)

	// Expect to see several resources listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "random_pet.pet[0]") &&
			strings.Contains(s, "random_pet.pet[1]") &&
			strings.Contains(s, "random_pet.pet[2]")
	})

	// Move resource random_pet.pet[0]
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Alt: true, Runes: []rune{'m'}})

	// Expect to see prompt prompting to enter a destination address
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Enter destination address: random_pet.pet[0]")
	})

	// Delete resource name pet[0] and replace with giraffe[99]
	tm.Send(tea.KeyMsg{Type: tea.KeyBackspace})
	tm.Send(tea.KeyMsg{Type: tea.KeyBackspace})
	tm.Send(tea.KeyMsg{Type: tea.KeyBackspace})
	tm.Send(tea.KeyMsg{Type: tea.KeyBackspace})
	tm.Send(tea.KeyMsg{Type: tea.KeyBackspace})
	tm.Send(tea.KeyMsg{Type: tea.KeyBackspace})
	tm.Type("giraffe[99]")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Expect to see new resource listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "random_pet.giraffe[99]")
	})
}

func TestWorkspace_Resources_Delete(t *testing.T) {
	t.Parallel()

	tm := setupResourcesTab(t)

	// Expect to see several resources listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "random_pet.pet[0]") &&
			strings.Contains(s, "random_pet.pet[1]") &&
			strings.Contains(s, "random_pet.pet[2]")
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
		return strings.Contains(s, "Delete 3 resource(s)? (y/N):")
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
}

func TestWorkspace_Resources_TargetedPlan(t *testing.T) {
	t.Parallel()

	tm := setupResourcesTab(t)

	// Expect to see several resources listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "random_pet.pet[0]") &&
			strings.Contains(s, "random_pet.pet[1]") &&
			strings.Contains(s, "random_pet.pet[2]")
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

func TestWorkspace_Resources_Reload(t *testing.T) {
	t.Parallel()

	tm := setupResourcesTab(t)

	// Remove several resource using a tool outside of pug.
	rm := exec.Command("terraform", "state", "rm", "random_pet.pet[0]", "random_pet.pet[1]")
	rm.Dir = filepath.Join(tm.workdir, "modules/a")
	err := rm.Run()
	require.NoError(t, err)

	// Reload state
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Expect reduced number of resources
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "resources (8)")
	})
}

func setupResourcesTab(t *testing.T) *testModel {
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
