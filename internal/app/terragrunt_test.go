package app

import (
	"os/exec"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTerragrunt_SingleInit(t *testing.T) {
	t.Parallel()
	skipIfTerragruntNotFound(t)

	_ = setupAndInitTerragruntModule(t)
}

func TestTerragrunt_SinglePlan(t *testing.T) {
	t.Parallel()
	skipIfTerragruntNotFound(t)

	tm := setupAndInitTerragruntModule(t)

	// Create plan on first module
	tm.Type("p")
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "Task.*plan.*default.*modules/a.*exited", s) &&
			strings.Contains(s, "terragrunt plan")

	})
}

func TestTerragrunt_SingleApply(t *testing.T) {
	t.Parallel()
	skipIfTerragruntNotFound(t)

	tm := setupAndInitTerragruntModule(t)

	// Create apply for module.
	tm.Type("a")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Auto-apply 1 modules? (y/N):")
	})
	tm.Type("y")

	// Send to apply task page
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `Task.*apply.*default.*modules/a.*\+10~0-0.*exited`, s) &&
			strings.Contains(s, "terragrunt apply")

	})
}

func skipIfTerragruntNotFound(t *testing.T) {
	if _, err := exec.LookPath("terragrunt"); err != nil {
		t.Skip("skipping test: terragrunt not found")
	}
}

func setupAndInitTerragruntModule(t *testing.T) *testModel {
	tm := setup(t, "./testdata/single_terragrunt_module", withProgram("terragrunt"))

	// Expect single module to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		t.Log(s)
		return matchPattern(t, "Task.*init.*modules/a.*exited", s)

	})

	// Show task info sidebar so tests can check that terragrunt is indeed being
	// executed.
	tm.Type("I")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "terragrunt init")

	})

	// Go back to modules listing
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Expect single modules to be listed, along with its default workspace.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "modules/a.*default", s)
	})

	return tm
}
