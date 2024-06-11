package workspace

import (
	"encoding/json"
	"fmt"
	"strings"

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
	return mm.MakePreview(res, width, height, tui.DefaultYPosition)
}

func (mm *ResourceMaker) MakePreview(res resource.Resource, width, height int, yPos int) (tea.Model, error) {
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
		viewport: tui.NewViewport(tui.ViewportOptions{
			Width:     width,
			Height:    height,
			YPosition: yPos,
			JSON:      true,
		}),
	}

	marshaled, err := json.Marshal(stateResource.Attributes)
	if err != nil {
		return nil, err
	}
	m.viewport.SetContent(string(marshaled), true)

	return m, nil
}

type resourceModel struct {
	viewport tui.Viewport
	resource *state.Resource
	height   int
	width    int
	borders  bool
	helpers  *tui.Helpers
}

func (m resourceModel) Init() tea.Cmd {
	return m.viewport.Init()
}

func (m resourceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

const (
	// bordersHeight is the total height of the borders to the top and
	// bottom of the content
	bordersHeight = 2
)

// View renders the viewport
func (m resourceModel) View() string {
	if m.borders {
		return fmt.Sprintf("%s\n%s\n%s",
			strings.Repeat("─", m.width),
			m.viewport.View(),
			strings.Repeat("─", m.width),
		)
	}
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
