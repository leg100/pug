package tui

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/resource"
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
	panes map[Position]ChildModel
	// pages tracks which page each pane is currently showing.
	pages map[Position]Page
	// total width and height of the terminal space available to panes.
	width, height int
	// leftPaneWidth is the width of the left pane when sharing the terminal
	// with other panes.
	leftPaneWidth int
	// topRightPaneHeight is the height of the top right pane.
	topRightHeight int
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
		panes:          make(map[Position]ChildModel),
		pages:          make(map[Position]Page),
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
		case key.Matches(msg, Keys.ShrinkPaneWidth):
			p.updateLeftWidth(-1)
			p.updateChildSizes()
		case key.Matches(msg, Keys.GrowPaneWidth):
			p.updateLeftWidth(1)
			p.updateChildSizes()
		case key.Matches(msg, Keys.ShrinkPaneHeight):
			p.updateTopRightHeight(-1)
			p.updateChildSizes()
		case key.Matches(msg, Keys.GrowPaneHeight):
			p.updateTopRightHeight(1)
			p.updateChildSizes()
		case key.Matches(msg, Keys.SwitchPane):
			p.cycleActivePane(false)
		case key.Matches(msg, Keys.SwitchPaneBack):
			p.cycleActivePane(true)
		case key.Matches(msg, Keys.ClosePane):
			cmds = append(cmds, p.closeActivePane())
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
	if table, ok := p.panes[TopRightPane].(tablePane); ok {
		if kind, id, ok := table.PreviewCurrentRow(); ok {
			cmd := p.setPane(NavigationMsg{
				Page:         Page{Kind: kind, ID: id},
				Position:     BottomRightPane,
				DisableFocus: true,
			})
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

// ActiveModel retrieves the model of the active pane.
func (p *PaneManager) ActiveModel() ChildModel {
	return p.panes[p.active]
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
	delete(p.pages, p.active)
	p.updateChildSizes()
	return p.cycleActivePane(false)
}

func (p *PaneManager) updateLeftWidth(delta int) {
	if p.panes[LeftPane] == nil {
		// There is no split to adjust
		return
	}
	p.leftPaneWidth = clamp(p.leftPaneWidth+delta, minPaneWidth, p.width-minPaneWidth)
}

func (p *PaneManager) updateTopRightHeight(delta int) {
	if p.panes[TopRightPane] == nil || p.panes[BottomRightPane] == nil {
		// There is no split to adjust
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
	return p.panes[position].Update(msg)
}

func (m *PaneManager) setPane(msg NavigationMsg) (cmd tea.Cmd) {
	if page, ok := m.pages[msg.Position]; ok && page == msg.Page {
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
	if msg.Position == TopRightPane {
		// A new top right pane replaces any bottom right pane as well.
		delete(m.panes, BottomRightPane)
	}
	m.panes[msg.Position] = model
	m.pages[msg.Position] = msg.Page
	m.updateChildSizes()
	if !msg.DisableFocus {
		focus := m.focusPane(msg.Position)
		cmd = tea.Batch(focus, cmd)
	}
	return cmd
}

func (m *PaneManager) focusPane(position Position) tea.Cmd {
	var cmds []tea.Cmd
	if previous, ok := m.panes[m.active]; ok {
		cmds = append(cmds, previous.Update(UnfocusPaneMsg{}))
	}
	m.active = position
	cmds = append(cmds, m.panes[m.active].Update(FocusPaneMsg{}))
	return tea.Batch(cmds...)
}

func (m *PaneManager) paneWidth(position Position) int {
	switch position {
	case LeftPane:
		if len(m.panes) > 1 {
			return m.leftPaneWidth
		}
	default:
		if m.panes[LeftPane] != nil {
			return max(minPaneWidth, m.width-m.leftPaneWidth)
		}
	}
	return m.width
}

func (m *PaneManager) paneHeight(position Position) int {
	switch position {
	case TopRightPane:
		if m.panes[BottomRightPane] != nil {
			return m.topRightHeight
		}
	case BottomRightPane:
		if m.panes[TopRightPane] != nil {
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

var border = map[bool]lipgloss.Border{
	true:  lipgloss.Border(lipgloss.ThickBorder()),
	false: lipgloss.Border(lipgloss.NormalBorder()),
}

func (m *PaneManager) renderPane(position Position) string {
	if m.panes[position] == nil {
		return ""
	}
	model := m.panes[position]
	isActive := position == m.active
	border := border[isActive]
	topBorder := m.buildTopBorder(position)
	renderedPane := lipgloss.NewStyle().
		Width(m.paneWidth(position) - 2).    // -2 for border
		Height(m.paneHeight(position) - 2).  // -2 for border
		MaxWidth(m.paneWidth(position) - 2). // -2 for border
		Render(model.View())
	return lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().Render(topBorder),
		lipgloss.NewStyle().
			BorderForeground(BorderColor(isActive)).
			Border(border, false, true, true, true).Render(renderedPane),
	)
}

func (m *PaneManager) buildTopBorder(position Position) string {
	var (
		model       = m.panes[position]
		width       = m.paneWidth(position)
		isActive    = position == m.active
		border      = border[isActive]
		borderStyle = lipgloss.NewStyle().Foreground(BorderColor(isActive))
	)
	var middle string
	if metadataModel, ok := model.(interface{ Metadata() string }); ok {
		// Render top border with metadata in the center
		//
		// total length of top border runes, not including corners
		length := max(0, width-2-lipgloss.Width(metadataModel.Metadata()))
		leftLength := length / 2
		rightLength := max(0, length-leftLength)
		renderedMetadata := metadataModel.Metadata()
		if isActive {
			// If active border, then strip any styling from metadata and apply
			// border style.
			renderedMetadata = internal.StripAnsi(renderedMetadata)
			renderedMetadata = borderStyle.Render(renderedMetadata)
		}
		middle = lipgloss.JoinHorizontal(lipgloss.Left,
			borderStyle.Render(strings.Repeat(border.Top, leftLength)),
			renderedMetadata,
			borderStyle.Render(strings.Repeat(border.Top, rightLength)),
		)
	} else {
		middle = borderStyle.Render(strings.Repeat(border.Top, max(0, width-2)))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left,
		borderStyle.Render(border.TopLeft),
		middle,
		borderStyle.Render(border.TopRight),
	)
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
