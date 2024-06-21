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

func TestState_SingleTaint_Untaint(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Taint first resource, which should be random_pet.pet[0]
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlT})

	// Expect to be taken to task page for taint
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Resource instance random_pet.pet[0] has been marked as tainted.")
	})

	// Go back to state page
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Expect resource to be marked as tainted
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "random_pet.pet[0] (tainted)")
	})

	// Untaint resource
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlU})

	// Expect to be taken to task page for untaint
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Resource instance random_pet.pet[0] has been successfully untainted.")
	})
}

func TestState_MultipleTaint_Untaint(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Taint all resources
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlT})

	// Expect to be taken to task group page for taint
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*taint.*10/10", s) &&
			matchPattern(t, `modules/a.*default.*exited`, s)
	})

	// Go back to state page
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Expect all resources to be marked as tainted. Note we wait until serial
	// 21 of the state has been loaded, which is the final state after 10 taints
	// (serial 11 + 10 = 21). If we didn't do this, then the select all
	// operation in the next step can be undone by a state reload, which loads
	// brand new resources, removing all selections.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `State.*21.*default.*modules/a`, s) &&
			strings.Contains(s, "random_pet.pet[0] (tainted)") &&
			strings.Contains(s, "random_pet.pet[1] (tainted)") &&
			strings.Contains(s, "random_pet.pet[2] (tainted)") &&
			strings.Contains(s, "random_pet.pet[3] (tainted)") &&
			strings.Contains(s, "random_pet.pet[4] (tainted)") &&
			strings.Contains(s, "random_pet.pet[5] (tainted)") &&
			strings.Contains(s, "random_pet.pet[6] (tainted)") &&
			strings.Contains(s, "random_pet.pet[7] (tainted)") &&
			strings.Contains(s, "random_pet.pet[8] (tainted)") &&
			strings.Contains(s, "random_pet.pet[9] (tainted)")
	})

	// Untaint all resources (need to select all again, because resources have
	// been reloaded).
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlU})

	// Expect to be taken to task group page for untaint
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*untaint.*10/10", s) &&
			matchPattern(t, `modules/a.*default.*exited`, s)
	})
}

func TestState_Move(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Move resource random_pet.pet[0]
	tm.Type("M")

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

	// Expect to be taken to task page for move
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `Task.*state mv.*default.*modules/a.*exited`, s) &&
			strings.Contains(s, `Move "random_pet.pet[0]" to "random_pet.giraffe[99]"`) &&
			strings.Contains(s, `Successfully moved 1 object(s).`)
	})
}

func TestState_SingleDelete(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Delete first resource
	tm.Type("D")

	// Confirm deletion
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Delete 1 resource(s)? (y/N):")
	})
	tm.Type("y")

	// User is taken to its task page, which should provide the output from the
	// command.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Removed random_pet.pet[0]") &&
			strings.Contains(s, "Successfully removed 1 resource instance(s).")
	})

	// Go back to state page
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Expect only 9 resources.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "1-9 of 9")
	})
}

func TestState_MultipleDelete(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Delete all resources
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("D")

	// Confirm deletion
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Delete 10 resource(s)? (y/N):")
	})
	tm.Type("y")

	// User is taken to its task page, which should provide the output from the
	// command.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Removed random_pet.pet[0]") &&
			strings.Contains(s, "Removed random_pet.pet[1]") &&
			strings.Contains(s, "Removed random_pet.pet[2]") &&
			strings.Contains(s, "Removed random_pet.pet[3]") &&
			strings.Contains(s, "Removed random_pet.pet[4]") &&
			strings.Contains(s, "Removed random_pet.pet[5]") &&
			strings.Contains(s, "Removed random_pet.pet[6]") &&
			strings.Contains(s, "Removed random_pet.pet[7]") &&
			strings.Contains(s, "Removed random_pet.pet[8]") &&
			strings.Contains(s, "Removed random_pet.pet[9]") &&
			strings.Contains(s, "Successfully removed 10 resource instance(s).")
	})

	// Go back to state page
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// TODO: test that there are zero resources in state. There is currently
	// scant information to test for.
}

func TestState_TargetedPlan_SingleResource(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Create targeted plan for first resource
	tm.Type("p")

	// Expect to be taken to the task page for the plan, with a completed plan, and a warning
	// that resource targeting is in effect
	waitFor(t, tm, func(s string) bool {
		// Strip ANSI formatting from output
		s = internal.StripAnsi(s)
		return matchPattern(t, `Task.*plan.*default.*modules/a.*\+1~0\-1.*exited`, s) &&
			strings.Contains(s, "Warning: Resource targeting is in effect")
	})
}

func TestState_TargetedPlan_MultipleResource(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Create targeted plan for all resources
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("p")

	// Expect to be taken to the task page for the plan, with a completed plan,
	// and a warning that resource targeting is in effect
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return matchPattern(t, `Task.*plan.*default.*modules/a.*\+10~0\-10.*exited`, s) &&
			strings.Contains(s, "Warning: Resource targeting is in effect")
	})
}

func TestState_Filter(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

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
		return strings.Contains(s, "1/10")
	})
}

func TestState_Reload(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Remove several resource using a tool outside of pug.
	rm := exec.Command("terraform", "state", "rm", "random_pet.pet[0]", "random_pet.pet[1]")
	rm.Dir = filepath.Join(tm.workdir, "modules/a")
	err := rm.Run()
	require.NoError(t, err)

	// Reload state
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Expect reduced number of resources
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "1-8 of 8")
	})
}

func TestState_NoState(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/workspace_resources_empty")

	// Wait for module to be loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!")
	})

	// Go workspaces
	tm.Type("w")

	// Wait for default workspace to be loaded
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `modules/a.*default.*✓.*0`, s)
	})

	// Go to state page
	tm.Type("s")

	// Expect message indicating no state found
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "No state found")
	})
}

func setupState(t *testing.T) *testModel {
	// Setup test with pre-existing state
	tm := setup(t, "./testdata/state")

	// Wait for module to be loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!")
	})

	// Go workspaces
	tm.Type("w")

	// Wait for default workspace to be loaded
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `modules/a.*default.*✓.*10`, s)
	})

	// Go to state page for workspace
	tm.Type("s")

	// Expect to see title and several resources listed
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `State.*11.*default.*modules/a`, s) &&
			strings.Contains(s, "random_pet.pet[0]") &&
			strings.Contains(s, "random_pet.pet[1]") &&
			strings.Contains(s, "random_pet.pet[2]") &&
			strings.Contains(s, "random_pet.pet[3]") &&
			strings.Contains(s, "random_pet.pet[4]") &&
			strings.Contains(s, "random_pet.pet[5]") &&
			strings.Contains(s, "random_pet.pet[6]") &&
			strings.Contains(s, "random_pet.pet[7]") &&
			strings.Contains(s, "random_pet.pet[8]") &&
			strings.Contains(s, "random_pet.pet[9]")
	})

	return tm
}
