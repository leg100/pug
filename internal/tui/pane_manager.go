package tui

import (
	"errors"
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui/keys"
	"golang.org/x/exp/maps"
)

type Position int

const (
	// TopRightPane occupies the top right area of the terminal. Mutually
	// exclusive with RightPane.
	TopRightPane Position = iota
	// BottomRightPane occupies the bottom right area of the terminal. Mutually
	// exclusive with RightPane.
	BottomRightPane
	// LeftPane occupies the left side of the terminal.
	LeftPane
)

// PaneManager manages the layout of the three panes that compose the Pug full screen terminal app.
type PaneManager struct {
	// makers for making models for panes
	makers map[Kind]Maker
	// cache of previously made models
	cache *Cache
	// the position of the currently active pane
	active Position
	// panes tracks currently visible panes
	panes map[Position]pane
	// total width and height of the terminal space available to panes.
	width, height int
	// leftPaneWidth is the width of the left pane when sharing the terminal
	// with other panes.
	leftPaneWidth int
	// topRightPaneHeight is the height of the top right pane.
	topRightHeight int
	// history tracks previously visited models for the top right pane.
	history []pane
}

type pane struct {
	model ChildModel
	page  Page
}

type tablePane interface {
	PreviewCurrentRow() (Kind, resource.ID, bool)
}

// NewPaneManager constructs the pane manager with at least the explorer, which
// occupies the left pane.
func NewPaneManager(makers map[Kind]Maker) *PaneManager {
	p := &PaneManager{
		makers:         makers,
		cache:          NewCache(),
		panes:          make(map[Position]pane),
		leftPaneWidth:  defaultLeftPaneWidth,
		topRightHeight: defaultTopRightPaneHeight,
	}
	return p
}

func (p *PaneManager) Init() tea.Cmd {
	return p.setPane(NavigationMsg{
		Position: LeftPane,
		Page:     Page{Kind: ExplorerKind},
	})
}

func (p *PaneManager) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Common.Back):
			if p.active != TopRightPane {
				// History is only maintained for the top right pane.
				break
			}
			if len(p.history) == 1 {
				// At dawn of history; can't go further back.
				return ReportError(errors.New("already at first page"))
			}
			// Pop current model from history
			p.history = p.history[:len(p.history)-1]
			// Set pane to last model
			p.panes[TopRightPane] = p.history[len(p.history)-1]
			// A new top right pane replaces any bottom right pane as well.
			delete(p.panes, BottomRightPane)
			p.updateChildSizes()
		case key.Matches(msg, keys.Global.ShrinkPaneWidth):
			p.updateLeftWidth(-1)
			p.updateChildSizes()
		case key.Matches(msg, keys.Global.GrowPaneWidth):
			p.updateLeftWidth(1)
			p.updateChildSizes()
		case key.Matches(msg, keys.Global.ShrinkPaneHeight):
			p.updateTopRightHeight(-1)
			p.updateChildSizes()
		case key.Matches(msg, keys.Global.GrowPaneHeight):
			p.updateTopRightHeight(1)
			p.updateChildSizes()
		case key.Matches(msg, keys.Navigation.SwitchPane):
			p.cycleActivePane(false)
		case key.Matches(msg, keys.Navigation.SwitchPaneBack):
			p.cycleActivePane(true)
		case key.Matches(msg, keys.Global.ClosePane):
			cmds = append(cmds, p.closeActivePane())
		case key.Matches(msg, keys.Navigation.LeftPane):
			cmds = append(cmds, p.focusPane(LeftPane))
		case key.Matches(msg, keys.Navigation.TopRightPane):
			cmds = append(cmds, p.focusPane(TopRightPane))
		case key.Matches(msg, keys.Navigation.BottomRightPane):
			cmds = append(cmds, p.focusPane(BottomRightPane))
		default:
			// Send remaining keys to active pane
			cmds = append(cmds, p.updateModel(p.active, msg))
		}
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.updateLeftWidth(0)
		p.updateTopRightHeight(0)
		p.updateChildSizes()
	case NavigationMsg:
		cmds = append(cmds, p.setPane(msg))
	default:
		// Send remaining message types to cached panes.
		cmds = p.cache.UpdateAll(msg)
	}

	// Check that if the top right pane is a table with a current row, then
	// ensure the bottom left pane corresponds to that current row, e.g. if the
	// top right pane is a tasks table, then the bottom right pane shows the
	// output for the current task row.
	if pane, ok := p.panes[TopRightPane]; ok {
		if table, ok := pane.model.(tablePane); ok {
			if kind, id, ok := table.PreviewCurrentRow(); ok {
				cmd := p.setPane(NavigationMsg{
					Page:         Page{Kind: kind, ID: id},
					Position:     BottomRightPane,
					DisableFocus: true,
				})
				cmds = append(cmds, cmd)
			}
		}
	}
	return tea.Batch(cmds...)
}

// ActiveModel retrieves the model of the active pane.
func (p *PaneManager) ActiveModel() ChildModel {
	return p.panes[p.active].model
}

// cycleActivePane makes the next pane the active pane. If last is true then the
// previous pane is made the active pane.
func (p *PaneManager) cycleActivePane(last bool) tea.Cmd {
	positions := maps.Keys(p.panes)
	slices.Sort(positions)
	var activeIndex int
	for i, pos := range positions {
		if pos == p.active {
			activeIndex = i
		}
	}
	var newActiveIndex int
	if last {
		newActiveIndex = activeIndex - 1
		if newActiveIndex < 0 {
			newActiveIndex = len(positions) + newActiveIndex
		}
	} else {
		newActiveIndex = (activeIndex + 1) % len(positions)
	}
	return p.focusPane(positions[newActiveIndex])
}

func (p *PaneManager) closeActivePane() tea.Cmd {
	if len(p.panes) == 1 {
		return ReportError(errors.New("cannot close last pane"))
	}
	delete(p.panes, p.active)
	p.updateChildSizes()
	return p.cycleActivePane(false)
}

func (p *PaneManager) updateLeftWidth(delta int) {
	if _, ok := p.panes[LeftPane]; !ok {
		// There is no vertical split to adjust
		return
	}
	p.leftPaneWidth = clamp(p.leftPaneWidth+delta, minPaneWidth, p.width-minPaneWidth)
}

func (p *PaneManager) updateTopRightHeight(delta int) {
	if _, ok := p.panes[TopRightPane]; !ok {
		// There is no horizontal split to adjust
		return
	} else if _, ok := p.panes[BottomRightPane]; !ok {
		// There is no horizontal split to adjust
		return
	}
	switch p.active {
	case BottomRightPane:
		delta = -delta
	}
	p.topRightHeight = clamp(p.topRightHeight+delta, minPaneHeight, p.height-minPaneHeight)
}

func (p *PaneManager) updateChildSizes() {
	for position := range p.panes {
		p.updateModel(position, tea.WindowSizeMsg{
			Width:  p.paneWidth(position) - 2,  // -2 for borders
			Height: p.paneHeight(position) - 2, // -2 for borders
		})
	}
}

func (p *PaneManager) updateModel(position Position, msg tea.Msg) tea.Cmd {
	return p.panes[position].model.Update(msg)
}

func (m *PaneManager) setPane(msg NavigationMsg) (cmd tea.Cmd) {
	if pane, ok := m.panes[msg.Position]; ok && pane.page == msg.Page {
		// Pane is already showing requested page, so just bring it into focus.
		if !msg.DisableFocus {
			return m.focusPane(msg.Position)
		}
		return nil
	}
	model := m.cache.Get(msg.Page)
	if model == nil {
		maker, ok := m.makers[msg.Page.Kind]
		if !ok {
			return ReportError(fmt.Errorf("no maker could be found for %s", msg.Page.Kind))
		}
		var err error
		model, err = maker.Make(msg.Page.ID, 0, 0)
		if err != nil {
			return ReportError(fmt.Errorf("making page of kind %s with id %s: %w", msg.Page.Kind, msg.Page.ID, err))
		}
		m.cache.Put(msg.Page, model)
		cmd = model.Init()
	}
	m.panes[msg.Position] = pane{
		model: model,
		page:  msg.Page,
	}
	if msg.Position == TopRightPane {
		// A new top right pane replaces any bottom right pane as well.
		delete(m.panes, BottomRightPane)
		// Track the models for the top right pane, so that the user can go back
		// to previous models.
		m.history = append(m.history, m.panes[TopRightPane])
	}
	m.updateChildSizes()
	if !msg.DisableFocus {
		focus := m.focusPane(msg.Position)
		cmd = tea.Batch(focus, cmd)
	}
	return cmd
}

func (m *PaneManager) focusPane(position Position) tea.Cmd {
	if _, ok := m.panes[position]; !ok {
		// There is no pane to focus at requested position
		return nil
	}
	var cmds []tea.Cmd
	if previous, ok := m.panes[m.active]; ok {
		cmds = append(cmds, previous.model.Update(UnfocusPaneMsg{}))
	}
	m.active = position
	cmds = append(cmds, m.panes[m.active].model.Update(FocusPaneMsg{}))
	return tea.Batch(cmds...)
}

func (m *PaneManager) paneWidth(position Position) int {
	switch position {
	case LeftPane:
		if len(m.panes) > 1 {
			return m.leftPaneWidth
		}
	default:
		if _, ok := m.panes[LeftPane]; ok {
			return max(minPaneWidth, m.width-m.leftPaneWidth)
		}
	}
	return m.width
}

func (m *PaneManager) paneHeight(position Position) int {
	switch position {
	case TopRightPane:
		if _, ok := m.panes[BottomRightPane]; ok {
			return m.topRightHeight
		}
	case BottomRightPane:
		if _, ok := m.panes[TopRightPane]; ok {
			return m.height - m.topRightHeight
		}
	}
	return m.height
}

func (m *PaneManager) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		removeEmptyStrings(
			m.renderPane(LeftPane),
			lipgloss.JoinVertical(lipgloss.Top,
				removeEmptyStrings(
					m.renderPane(TopRightPane),
					m.renderPane(BottomRightPane),
				)...,
			),
		)...,
	)
}

func (m *PaneManager) renderPane(position Position) string {
	if _, ok := m.panes[position]; !ok {
		return ""
	}
	model := m.panes[position].model
	isActive := position == m.active
	renderedPane := lipgloss.NewStyle().
		Width(m.paneWidth(position) - 2).    // -2 for border
		Height(m.paneHeight(position) - 2).  // -2 for border
		MaxWidth(m.paneWidth(position) - 2). // -2 for border
		Render(model.View())
	// Optionally, the pane model can embed text in its borders.
	borderTexts := make(map[BorderPosition]string)
	if textInBorder, ok := model.(interface {
		BorderText() map[BorderPosition]string
	}); ok {
		borderTexts = textInBorder.BorderText()
	}
	if !isActive {
		switch position {
		case LeftPane:
			borderTexts[TopRightBorder] = keys.Navigation.LeftPane.Keys()[0]
		case TopRightPane:
			borderTexts[TopRightBorder] = keys.Navigation.TopRightPane.Keys()[0]
		case BottomRightPane:
			borderTexts[TopRightBorder] = keys.Navigation.BottomRightPane.Keys()[0]
		}
	}
	return borderize(renderedPane, isActive, borderTexts)
}

func (m *PaneManager) HelpBindings() (bindings []key.Binding) {
	if m.active == TopRightPane {
		// Only the top right pane has the ability to "go back"
		bindings = append(bindings, keys.Common.Back)
	}
	if model, ok := m.ActiveModel().(ModelHelpBindings); ok {
		bindings = append(bindings, model.HelpBindings()...)
	}
	return bindings
}

func removeEmptyStrings(strs ...string) []string {
	n := 0
	for _, s := range strs {
		if s != "" {
			strs[n] = s
			n++
		}
	}
	return strs[:n]
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
