package app

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceList_SetCurrentWorkspace(t *testing.T) {
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

	// Go to global workspaces page
	tm.Type("W")

	// Expect two workspaces to be listed, and expect default to be the current
	// workspace
	waitFor(t, tm, func(s string) bool {
		defaultIsCurrent, err := regexp.MatchString(`default.*✓`, s)
		require.NoError(t, err)
		return defaultIsCurrent && strings.Contains(s, "dev")
	})

	// Navigate one row down to the dev workspace (default should be the first
	// row because it is lexicographically before dev).
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	// Make dev the current workspace
	tm.Type("C")

	// Expect dev to be the new current workspace
	waitFor(t, tm, func(s string) bool {
		devIsCurrent, err := regexp.MatchString(`dev.*✓`, s)
		require.NoError(t, err)
		return devIsCurrent
	})
}

func TestWorkspaceList_CreateRun(t *testing.T) {
	tm := setup(t)

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

	// Create run on first workspace
	tm.Type("p")

	// Expect to be taken to the run's page
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Run[default](modules/a)")
	})

	// Return to global workspaces page
	tm.Type("W")

	// Wait for all four workspaces to be listed
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `(?s)WORKSPACE.*default.*dev.*default.*default`, s)
	})

	// Create run on all four workspaces
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("p")

	// Expect all four workspaces' current run to enter the planned state
	waitFor(t, tm, func(s string) bool {
		return strings.Count(s, "planned") == 4
	})

}

func TestWorkspaceList_ApplyCurrentRun(t *testing.T) {
	tm := setup(t)

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

	// Attempt to apply first workspace
	tm.Type("a")

	// Expect error message because workspace does not have current run, yet.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "Error:.*workspace does not have a current run", s)
	})

	// Create run on all four workspaces
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("p")

	// Expect all four workspaces' current run to enter the planned state
	waitFor(t, tm, func(s string) bool {
		return strings.Count(s, "planned") == 4
	})

	// Apply all four workspaces
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("a")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply 4 runs (y/N)?")
	})
	tm.Type("y")

	// Expect all four workspaces' current run to enter the applied state
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `(?s)default.*applied.*dev.*applied.*default.*applied.*default.*applied`, s)
	})
}
