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

	// Expect short message in footer.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "taint: finished successfully…(Press 'o' for full output)") &&
			strings.Contains(s, "random_pet.pet[0] (tainted)")
	})

	// Untaint resource
	tm.Type("U")

	// Expect short message in footer.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "untaint: finished successfully…(Press 'o' for full output)") &&
			strings.Contains(s, "random_pet.pet[0]")
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
		return strings.Contains(s, "taint 10/10") &&
			matchPattern(t, `modules/a.*default.*taint.*exited`, s)
	})

	// Go back to state page
	tm.Type("s")

	// Expect all resources to be marked as tainted. Note we wait until serial
	// 21 of the state has been loaded, which is the final state after 10 taints
	// (serial 11 + 10 = 21). If we didn't do this, then the select all
	// operation in the next step can be undone by a state reload, which loads
	// brand new resources, removing all selections.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "state 󰠱 modules/a  default") &&
			strings.Contains(s, "random_pet.pet[0] (tainted)") &&
			strings.Contains(s, "random_pet.pet[1] (tainted)") &&
			strings.Contains(s, "random_pet.pet[2] (tainted)") &&
			strings.Contains(s, "random_pet.pet[3] (tainted)") &&
			strings.Contains(s, "random_pet.pet[4] (tainted)") &&
			strings.Contains(s, "random_pet.pet[5] (tainted)") &&
			strings.Contains(s, "random_pet.pet[6] (tainted)") &&
			strings.Contains(s, "random_pet.pet[7] (tainted)") &&
			strings.Contains(s, "random_pet.pet[8] (tainted)") &&
			strings.Contains(s, "random_pet.pet[9] (tainted)") &&
			strings.Contains(s, "#21")
	})

	// Untaint all resources (need to select all again, because resources have
	// been reloaded).
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("U")

	// Expect to be taken to task group page for untaint
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "untaint 10/10") &&
			matchPattern(t, `modules/a.*default.*untaint.*exited`, s)
	})
}

func TestState_Move(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Move resource random_pet.pet[0]
	tm.Type("m")

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

	// Expect short message in footer.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "random_pet.giraffe[99]") &&
			strings.Contains(s, "state mv: finished successfully…(Press 'o' for full output)")
	})
}

func TestState_SingleDelete(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Delete first resource
	tm.Send(tea.KeyMsg{Type: tea.KeyDelete})

	// Confirm deletion
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Delete 1 resource(s)? (y/N):")
	})
	tm.Type("y")

	// Expect short message in footer.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "state rm: finished successfully…(Press 'o' for full output)") &&
			// Expect only 9 resources now.
			strings.Contains(s, "1-9 of 9")
	})
}

func TestState_MultipleDelete(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Delete all resources
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Send(tea.KeyMsg{Type: tea.KeyDelete})

	// Confirm deletion
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Delete 10 resource(s)? (y/N):")
	})
	tm.Type("y")

	// User is taken to its task page, which should provide the output from the
	// command.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "state rm: finished successfully…(Press 'o' for full output)") &&
			// Expect only 0 resources now.
			strings.Contains(s, "1-0 of 0")
	})
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
		return strings.Contains(s, "plan 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +1~0-1") &&
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
		return strings.Contains(s, "plan 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +10~0-10") &&
			strings.Contains(s, "Warning: Resource targeting is in effect")
	})
}

func TestState_TargetedPlanDestroy_SingleResource(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Create targeted destroy plan for first resource
	tm.Type("d")

	// Expect to be taken to the task page for the destroy plan, with a
	// completed destroy plan, and a warning that resource targeting is in
	// effect
	waitFor(t, tm, func(s string) bool {
		// Strip ANSI formatting from output
		s = internal.StripAnsi(s)
		return strings.Contains(s, "plan (destroy) 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +0~0-1") &&
			strings.Contains(s, "Warning: Resource targeting is in effect")
	})
}

func TestState_TargetedPlanDestroy_MultipleResource(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Create targeted destroy plan for all resources
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("d")

	// Expect to be taken to the task page for the destroy plan, with a
	// completed destroy plan, and a warning that resource targeting is in
	// effect
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return strings.Contains(s, "plan (destroy) 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +0~0-10") &&
			strings.Contains(s, "Warning: Resource targeting is in effect")
	})
}

func TestState_TargetedApply_SingleResource(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Create targeted apply for first resource
	tm.Type("a")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Auto-apply 1 resources? (y/N):")
	})
	tm.Type("y")

	// Expect to be taken to the task page for the apply, with a completed apply, and a warning
	// that resource targeting is in effect
	waitFor(t, tm, func(s string) bool {
		// Strip ANSI formatting from output
		s = internal.StripAnsi(s)
		return strings.Contains(s, "apply 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +1~0-1") &&
			strings.Contains(s, "Warning: Resource targeting is in effect")
	})
}

func TestState_TargetedApply_MultipleResource(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Create targeted apply for all resources
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("a")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Auto-apply 10 resources? (y/N):")
	})
	tm.Type("y")

	// Expect to be taken to the task page for the apply, with a completed apply,
	// and a warning that resource targeting is in effect
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return strings.Contains(s, "apply 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +10~0-10") &&
			strings.Contains(s, "Warning: Resource targeting is in effect")
	})
}

func TestState_TargetedDestroy_SingleResource(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Create targeted destroy for first resource
	tm.Type("D")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Destroy 1 resources? (y/N):")
	})
	tm.Type("y")

	// Expect to be taken to the task page for the destroy, with a completed destroy, and a warning
	// that resource targeting is in effect
	waitFor(t, tm, func(s string) bool {
		// Strip ANSI formatting from output
		s = internal.StripAnsi(s)
		return strings.Contains(s, "apply (destroy) 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +0~0-1") &&
			strings.Contains(s, "Warning: Resource targeting is in effect")
	})
}

func TestState_TargetedDestroy_MultipleResource(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Create targeted destroy for all resources
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("D")

	// Give approval
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Destroy 10 resources? (y/N):")
	})
	tm.Type("y")

	// Expect to be taken to the task page for the destroy, with a completed destroy,
	// and a warning that resource targeting is in effect
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return strings.Contains(s, "apply (destroy) 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +0~0-10") &&
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

	// Expect single module to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "└ 󰠱 a")
	})

	// Initialize module
	tm.Type("i")
	// Expect init to succeed, and to populate pug with one workspace with 10
	// resources
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!") &&
			strings.Contains(s, "init 󰠱 modules/a") &&
			strings.Contains(s, "exited") &&
			strings.Contains(s, "└  default 0")
	})

	// Go to state page for workspace
	tm.Type("s")

	// Expect message indicating no state found
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "No state found")
	})
}

func TestState_ViewResource(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Press enter to view resource, which should be random_pet.pet[0]
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, `resource random_pet.pet[0]`)
	})
}

func TestState_ViewResourceTargetPlan(t *testing.T) {
	t.Parallel()

	tm := setupState(t)

	// Press enter to view resource, which should be random_pet.pet[0]
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, `resource random_pet.pet[0]`)
	})

	// Create targeted plan for resource
	tm.Type("p")

	// Expect to be taken to the task page for the plan, with a completed plan, and a warning
	// that resource targeting is in effect
	waitFor(t, tm, func(s string) bool {
		// Strip ANSI formatting from output
		s = internal.StripAnsi(s)
		return strings.Contains(s, "plan 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +1~0-1") &&
			strings.Contains(s, "Warning: Resource targeting is in effect")
	})
}

func setupState(t *testing.T) *testModel {
	// Setup test with pre-existing state
	tm := setup(t, "./testdata/state")

	// Expect single module to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "└ 󰠱 a")
	})

	// Initialize module
	tm.Type("i")
	// Expect init to succeed, and to populate pug with one workspace with 10
	// resources
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!") &&
			strings.Contains(s, "init 󰠱 modules/a") &&
			strings.Contains(s, "exited") &&
			strings.Contains(s, "└  default 10")
	})

	// Go to state page for workspace
	tm.Type("s")

	// Expect to see title and several resources listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "state 󰠱 modules/a  default") &&
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
