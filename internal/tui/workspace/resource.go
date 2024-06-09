package workspace

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hokaccha/go-prettyjson"
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
		height:   height,
		width:    width,
		helpers:  mm.Helpers,
		borders:  !mm.disableBorders,
		resource: stateResource,
	}

	// marshal resource's attributes to json and set as the viewport's content.
	marshaled, err := json.MarshalIndent(stateResource.Attributes, "", "\t")
	if err != nil {
		return nil, err
	}
	prettified, err := prettyjson.Format([]byte(marshaled))
	if err != nil {
		return nil, err
	}
	m.viewport = viewport.New(0, 0)
	m.viewport.SetContent(string(prettified))

	m.setWidth(width)
	m.setHeight(height)

	return m, nil
}

type resourceModel struct {
	viewport viewport.Model
	resource *state.Resource
	height   int
	width    int
	borders  bool
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
		m.setWidth(msg.Width)
		m.setHeight(msg.Height)
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

const (
	// scrollPercentWidth is the width of the scroll percentage section to the
	// right of the viewport
	scrollPercentWidth = 10
	// bordersWidth is the total width of the borders to the left and
	// right of the content
	bordersWidth = 2
	// bordersHeight is the total height of the borders to the top and
	// bottom of the content
	bordersHeight = 2
)

var borderStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder())

func (m *resourceModel) setWidth(width int) {
	m.width = width

	viewportWidth := width - scrollPercentWidth
	if m.borders {
		viewportWidth -= bordersWidth
	}
	m.viewport.Width = max(0, viewportWidth)
}

func (m *resourceModel) setHeight(height int) {
	if m.borders {
		height -= bordersHeight
	}
	m.viewport.Height = height
	m.height = height
}

// View renders the viewport
func (m resourceModel) View() string {
	var components []string

	viewport := tui.Regular.Copy().
		MaxWidth(m.viewport.Width).
		Render(m.viewport.View())
	components = append(components, viewport)

	// scroll percent container occupies a fixed width section to the right of
	// the viewport.
	scrollPercent := tui.Regular.Copy().
		Background(tui.ScrollPercentageBackground).
		Padding(0, 1).
		Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	scrollPercentContainer := tui.Regular.Copy().
		Margin(0, 1).
		Height(m.height).
		// subtract 2 to account for margins
		Width(scrollPercentWidth - 2).
		AlignVertical(lipgloss.Bottom).
		Render(scrollPercent)
	components = append(components, scrollPercentContainer)

	content := lipgloss.JoinHorizontal(lipgloss.Left, components...)

	if m.borders {
		return borderStyle.Render(content)
	}
	return content
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
