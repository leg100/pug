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
			strings.Contains(s, "terragrunt")

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
			strings.Contains(s, "terragrunt")

	})
}

// TestTerragrunt_Dependencies tests that terragrunt dependencies are
// respected.
func TestTerragrunt_Dependencies(t *testing.T) {
	t.Parallel()
	skipIfTerragruntNotFound(t)

	tm := setupAndInitTerragruntModulesWithDependencies(t)

	// Select all modules and create apply on each
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("a")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Auto-apply 6 modules? (y/N):")
	})
	tm.Type("y")

	// Expect 6 applies. The "." module fails because it doesn't have any config
	// files.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*apply.*1/5/6", s) &&
			matchPattern(t, `modules/vpc.*default.*\+0~0-0`, s) &&
			matchPattern(t, `modules/redis.*default.*\+0~0-0`, s) &&
			matchPattern(t, `modules/mysql.*default.*\+0~0-0`, s) &&
			matchPattern(t, `modules/backend-app.*default.*\+0~0-0`, s) &&
			matchPattern(t, `modules/frontend-app.*default.*\+0~0-0`, s) &&
			matchPattern(t, `\..*default.*errored`, s)
	})

	// Go back to modules listing
	tm.Type("m")

	// Expect several modules to now have some resources
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "modules/vpc.*default.*0", s) &&
			matchPattern(t, `modules/mysql.*default.*1`, s) &&
			matchPattern(t, "modules/redis.*default.*1", s) &&
			matchPattern(t, "modules/backend-app.*default.*3", s) &&
			matchPattern(t, "modules/frontend-app.*default.*2", s)
	})

	// Destroy resources in all modules (they should still all be selected).
	tm.Type("d")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Destroy resources of 6 modules? (y/N):")
	})
	tm.Type("y")

	// Expect 6 apply tasks.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*apply.*6/6", s) &&
			matchPattern(t, `modules/vpc.*default.*\+0~0-0`, s) &&
			matchPattern(t, `modules/redis.*default.*\+0~0-0`, s) &&
			matchPattern(t, `modules/mysql.*default.*\+0~0-0`, s) &&
			matchPattern(t, `modules/backend-app.*default.*\+0~0-0`, s) &&
			matchPattern(t, `modules/frontend-app.*default.*\+0~0-0`, s) &&
			matchPattern(t, `\..*default.*exited`, s)
	})

	// Go to modules listing
	tm.Type("m")

	// Expect modules to now have some 0 resources
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "modules/vpc.*default.*0", s) &&
			matchPattern(t, `modules/mysql.*default.*0`, s) &&
			matchPattern(t, "modules/redis.*default.*0", s) &&
			matchPattern(t, "modules/backend-app.*default.*0", s) &&
			matchPattern(t, "modules/frontend-app.*default.*0", s)
	})
}

func skipIfTerragruntNotFound(t *testing.T) {
	if _, err := exec.LookPath("terragrunt"); err != nil {
		t.Skip("skipping test: terragrunt not found")
	}
}

func setupAndInitTerragruntModule(t *testing.T) *testModel {
	tm := setup(t, "./testdata/single_terragrunt_module", withTerragrunt())

	// Expect single module to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "Task.*init.*modules/a.*exited", s)

	})

	// Show task info sidebar so tests can check that terragrunt is indeed being
	// executed.
	tm.Type("I")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "terragrunt")

	})

	// Go back to modules listing
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Expect single modules to be listed, along with its default workspace.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "modules/a.*default", s)
	})

	return tm
}

func setupAndInitTerragruntModulesWithDependencies(t *testing.T) *testModel {
	tm := setup(t, "./testdata/terragrunt_modules_with_dependencies", withTerragrunt())

	// Expect several modules to be listed, along with dependencies.
	//
	// NOTE: the integration test terminal width is limited, therefore
	// the dependencies column is cut short with an ellipsis, preventing full
	// assertion of all dependencies.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/vpc") &&
			matchPattern(t, `modules/mysql.*modules/vpc`, s) &&
			matchPattern(t, `modules/redis.*modules/vpc`, s) &&
			strings.Contains(s, `modules/backend-app`) &&
			strings.Contains(s, `modules/frontend-app`) &&
			matchPattern(t, `\..*local.*default`, s)
	})

	// Select all modules and init
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*init", s) &&
			matchPattern(t, `modules/vpc.*init.*exited`, s) &&
			matchPattern(t, `modules/redis.*init.*exited`, s) &&
			matchPattern(t, `modules/mysql.*init.*exited`, s) &&
			matchPattern(t, `modules/frontend-app.*init.*exited`, s) &&
			matchPattern(t, `modules/backend-app.*init.*exited`, s) &&
			matchPattern(t, `\..*init.*exited`, s)
	})

	// Go to workspace listing and expect one workspace for each module. We make
	// assertions here on the workspace listing rather than the module listing
	// because the latter has a dependencies column which makes it tricky to
	// write regular expressions that don't accidentally pick up false
	// positives...
	tm.Type("w")

	// Expect modules to be listed along with their default workspace.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `modules/vpc.*default`, s) &&
			matchPattern(t, `modules/mysql.*default`, s) &&
			matchPattern(t, `modules/redis.*default`, s) &&
			matchPattern(t, `modules/backend-app.*default`, s) &&
			matchPattern(t, `modules/frontend-app.*default`, s) &&
			matchPattern(t, `\..*default`, s)
	})

	// Go to modules listing and clear selection
	tm.Type("m")
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlBackslash})

	return tm
}
