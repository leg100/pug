package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal"
)

func TestModule(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Expect module to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")

	// Expect to see successful init message
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!")
	})
}

func TestModule_SetCurrentWorkspace(t *testing.T) {
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

	// Go to module A's page
	tm.Type("m")

	// Expect two workspaces to be listed, and expect default to be the current
	// workspace
	waitFor(t, tm, func(s string) bool {
		defaultIsCurrent := matchPattern(t, `default.*✓`, s)
		return defaultIsCurrent && strings.Contains(s, "dev")
	})

	// Navigate one row down to the dev workspace (default should be the first
	// row because it is lexicographically before dev).
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	// Make dev the current workspace
	tm.Type("C")

	// Expect dev to be the new current workspace
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `dev.*✓`, s)
	})
}

func TestModule_ReloadWorkspaces(t *testing.T) {
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

	// Go to module's page
	tm.Type("m")

	// Expect two workspaces to be listed
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `(?s)WORKSPACE.*default.*dev`, s)
	})

	// Reload workspaces for current module
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlW})

	// Expect message to inform user that reload has finished.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "completed reload-workspaces task successfully")
	})
}

// TestModule_Destroy demonstrates creating a destroy plan on a module.
func TestModule_Destroy(t *testing.T) {
	t.Parallel()

	// Setup test with pre-existing state
	tm := setup(t, "./testdata/module_destroy", keepState())

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

	// Wait for the one-and-only default workspace - which should be the current
	// workspace - and which should have 10 resources in its state loaded.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `default.*✓.*10`, s)
	})

	// Create destroy plan
	tm.Type("d")

	// Expect umpteen resources to be proposed for deletion
	waitFor(t, tm, func(s string) bool {
		t.Log(s)
		s = internal.StripAnsi(s)
		return strings.Contains(s, "+0~0-10")
	})
}
