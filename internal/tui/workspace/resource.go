package workspace

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
)

type ResourceMaker struct {
	Helpers *tui.Helpers

	disableBorders bool
}

func (mm *ResourceMaker) Make(res resource.Resource, width, height int) (tea.Model, error) {
	stateResource, ok := res.(*state.Resource)
	if !ok {
		return nil, fmt.Errorf("constructing state resource model: unexpected resource type: %T", res)
	}

	m := resourceModel{
		helpers:  mm.Helpers,
		resource: stateResource,
	}

	marshaled, err := json.MarshalIndent(stateResource.Attributes, "", "\t")
	if err != nil {
		return nil, err
	}
	m.viewport = tui.NewViewport(tui.ViewportOptions{
		Width:  width,
		Height: height,
		JSON:   true,
		Border: !mm.disableBorders,
	})
	m.viewport.AppendContent(string(marshaled), true)

	return m, nil
}

type resourceModel struct {
	viewport tui.Viewport
	resource *state.Resource
	helpers  *tui.Helpers
}

func (m resourceModel) Init() tea.Cmd {
	return nil
}

func (m resourceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport, _ = m.viewport.Update(tea.WindowSizeMsg{
			Width:  msg.Width,
			Height: msg.Height,
		})
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m resourceModel) View() string {
	return m.viewport.View()
}

func (m resourceModel) Title() string {
	return m.helpers.Breadcrumbs("Resource", m.resource)
}

func (m resourceModel) HelpBindings() []key.Binding {
	bindings := []key.Binding{
		keys.Common.Cancel,
	}
	return bindings
}
