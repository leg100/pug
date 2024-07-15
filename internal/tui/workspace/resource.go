package workspace

import (
	"encoding/json"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
)

type ResourceMaker struct {
	StateService tui.StateService
	RunService   tui.RunService
	Helpers      *tui.Helpers

	disableBorders bool
}

func (mm *ResourceMaker) Make(id resource.ID, width, height int) (tea.Model, error) {
	stateResource, err := mm.StateService.GetResource(id)
	if err != nil {
		return nil, err
	}

	m := resourceModel{
		StateService: mm.StateService,
		RunService:   mm.RunService,
		helpers:      mm.Helpers,
		resource:     stateResource,
		border:       !mm.disableBorders,
	}

	marshaled, err := json.MarshalIndent(stateResource.Attributes, "", "\t")
	if err != nil {
		return nil, err
	}
	m.viewport = tui.NewViewport(tui.ViewportOptions{
		Width:  m.viewportWidth(width),
		Height: m.viewportHeight(height),
		JSON:   true,
	})
	m.viewport.AppendContent(string(marshaled), true)

	return m, nil
}

type resourceModel struct {
	StateService tui.StateService
	RunService   tui.RunService

	viewport tui.Viewport
	resource *state.Resource
	helpers  *tui.Helpers
	border   bool
}

func (m resourceModel) Init() tea.Cmd {
	return nil
}

func (m resourceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd              tea.Cmd
		cmds             []tea.Cmd
		createRunOptions run.CreateOptions
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, resourcesKeys.Taint):
			fn := func(workspaceID resource.ID) (*task.Task, error) {
				return m.StateService.Taint(workspaceID, m.resource.Address)
			}
			return m, m.helpers.CreateTasks("taint", fn, m.resource.Workspace().GetID())
		case key.Matches(msg, resourcesKeys.Untaint):
			fn := func(workspaceID resource.ID) (*task.Task, error) {
				return m.StateService.Untaint(workspaceID, m.resource.Address)
			}
			return m, m.helpers.CreateTasks("untaint", fn, m.resource.Workspace().GetID())
		case key.Matches(msg, resourcesKeys.Move):
			return m, m.helpers.Move(m.resource.Workspace().GetID(), m.resource.Address)
		case key.Matches(msg, keys.Common.Delete):
			fn := func(workspaceID resource.ID) (*task.Task, error) {
				return m.StateService.Delete(workspaceID, m.resource.Address)
			}
			return m, tui.YesNoPrompt(
				"Delete resource?",
				m.helpers.CreateTasks("state-rm", fn, m.resource.Workspace().GetID()),
			)
		case key.Matches(msg, keys.Common.PlanDestroy):
			// Create a targeted destroy plan.
			createRunOptions.Destroy = true
			fallthrough
		case key.Matches(msg, keys.Common.Plan):
			// Create a targeted plan.
			createRunOptions.TargetAddrs = []state.ResourceAddress{m.resource.Address}
			fn := func(workspaceID resource.ID) (*task.Task, error) {
				return m.RunService.Plan(workspaceID, createRunOptions)
			}
			return m, m.helpers.CreateTasks("plan", fn, m.resource.Workspace().GetID())
		}
	case tea.WindowSizeMsg:
		m.viewport.SetDimensions(m.viewportWidth(msg.Width), m.viewportHeight(msg.Height))
		return m, nil
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m resourceModel) View() string {
	if m.border {
		return tui.Border.Render(m.viewport.View())
	}
	return m.viewport.View()
}

func (m resourceModel) viewportWidth(width int) int {
	if m.border {
		width -= 2
	}
	return max(0, width)
}

func (m resourceModel) viewportHeight(height int) int {
	if m.border {
		height -= 2
	}
	return max(0, height)
}

func (m resourceModel) Title() string {
	var tainted string
	if m.resource.Tainted {
		tainted = tui.TitleTainted.Render("tainted")
	}
	return tui.Breadcrumbs("Resource", m.resource) + tainted
}

func (m resourceModel) HelpBindings() []key.Binding {
	return []key.Binding{
		keys.Common.Plan,
		keys.Common.PlanDestroy,
		keys.Common.Delete,
		resourcesKeys.Move,
		resourcesKeys.Taint,
		resourcesKeys.Untaint,
	}
}
