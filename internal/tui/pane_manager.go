package tui

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/exp/maps"
)

const (
	// default height of the top right pane when split, including borders.
	defaultTopRightHeight = 15
	// minimum height of the top right pane, including borders.
	minTopRightHeight = 10
	// minimum height of the bottom right pane, including borders.
	minBottomRightHeight = MinContentHeight - minTopRightHeight
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
	// total width and height of the terminal space available to panes.
	width, height int
	// minimum width and heights for panes
	minWidth, minHeight int
	// leftPaneWidth is the width of the left pane when sharing the terminal
	// with other panes.
	leftPaneWidth int
	// topRightPaneHeight is the height of the top right pane.
	topRightHeight int
}

type pane struct {
	model  ChildModel
	width  int
	height int
}

// NewPaneManager constructs the pane manager with at least the explorer, which
// occupies the left pane.
func NewPaneManager(explorer ChildModel, makers map[Kind]Maker) *PaneManager {
	cache := NewCache()
	cache.Put(Page{Kind: ExplorerKind}, explorer)

	p := &PaneManager{
		makers: makers,
		cache:  cache,
		active: LeftPane,
		panes: map[Position]ChildModel{
			LeftPane: explorer,
		},
		minWidth:       40,
		minHeight:      10,
		leftPaneWidth:  30,
		topRightHeight: defaultTopRightHeight,
	}
	return p
}

func (p *PaneManager) Init() tea.Cmd {
	return p.panes[p.active].Init()
}

func (p *PaneManager) Update(msg tea.Msg) tea.Cmd {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
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
			p.cycleActivePane()
		case key.Matches(msg, Keys.ClosePane):
			if err := p.closeActivePane(); err != nil {
				cmd = ReportError(err)
			}
		default:
			// Send remaining keys to active pane
			cmd = p.updateModel(p.active, msg)
		}
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.updateLeftWidth(0)
		p.updateTopRightHeight(0)
		p.updateChildSizes()
	default:
		// Send remaining message types to active pane
		cmd = p.updateModel(p.active, msg)
	}
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

// ActiveModel retrieves the model of the active pane.
func (p *PaneManager) ActiveModel() ChildModel {
	return p.panes[p.active]
}

func (p *PaneManager) cycleActivePane() {
	positions := maps.Keys(p.panes)
	slices.Sort(positions)
	var activeIndex int
	for i, pos := range positions {
		if pos == p.active {
			activeIndex = i
		}
	}
	newActiveIndex := (activeIndex + 1) % len(positions)
	p.active = positions[newActiveIndex]
}

func (p *PaneManager) closeActivePane() error {
	if len(p.panes) == 1 {
		return errors.New("cannot close last pane")
	}
	delete(p.panes, p.active)
	p.cycleActivePane()
	p.updateChildSizes()
	return nil
}

func (p *PaneManager) updateLeftWidth(delta int) {
	if p.panes[LeftPane] == nil {
		// There is no split to adjust
		return
	}
	p.leftPaneWidth = clamp(p.leftPaneWidth+delta, p.minWidth, p.width-p.minWidth)
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
	p.topRightHeight = clamp(p.topRightHeight+delta, minTopRightHeight, p.height-minBottomRightHeight)
}

func (p *PaneManager) updateChildSizes() {
	for position := range p.panes {
		p.updateModel(position, tea.WindowSizeMsg{
			Width:  p.paneWidth(position) - 2,
			Height: p.paneHeight(position) - 2,
		})
	}
}

func (p *PaneManager) updateModel(position Position, msg tea.Msg) tea.Cmd {
	return p.panes[position].Update(msg)
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

func (m *PaneManager) SetPane(page Page, position Position) (cmd tea.Cmd, err error) {
	model := m.cache.Get(page)
	if model == nil {
		maker, ok := m.makers[page.Kind]
		if !ok {
			return nil, fmt.Errorf("no maker could be found for %s", page.Kind)
		}
		model, err = maker.Make(page.ID, 0, 0)
		if err != nil {
			return nil, fmt.Errorf("making page: %w", err)
		}
		m.cache.Put(page, model)
		cmd = model.Init()
	}
	switch position {
	case TopRightPane:
		// TopRightPane replaces BottomRightPane too
		delete(m.panes, BottomRightPane)
	}
	m.active = position
	m.panes[position] = model
	m.updateChildSizes()
	return cmd, nil
}

func (m *PaneManager) paneWidth(position Position) int {
	switch position {
	case LeftPane:
		if len(m.panes) > 1 {
			return m.leftPaneWidth
		}
	default:
		if m.panes[LeftPane] != nil {
			return max(m.minWidth, m.width-m.leftPaneWidth)
		}
	}
	return m.width
}

func (m *PaneManager) paneHeight(position Position) int {
	switch position {
	case TopRightPane:
		if m.panes[BottomRightPane] != nil {
			return max(minTopRightHeight, m.topRightHeight)
		}
	case BottomRightPane:
		if m.panes[TopRightPane] != nil {
			return max(minBottomRightHeight, m.height-m.topRightHeight)
		}
	}
	return m.height
}

func (m *PaneManager) renderPane(position Position) string {
	if m.panes[position] == nil {
		return ""
	}
	model := m.panes[position]
	renderedPane := model.View()
	isActive := position == m.active
	border := border[isActive]
	topBorder := m.buildTopBorder(position)
	return lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().Render(topBorder),
		lipgloss.NewStyle().
			BorderForeground(BorderColor(isActive)).
			Border(border, false, true, true, true).Render(renderedPane),
	)
}

func (m *PaneManager) buildTopBorder(position Position) string {
	var (
		model    = m.panes[position]
		width    = m.paneWidth(position)
		isActive = position == m.active
		border   = border[isActive]
	)
	var middle string
	if metadataModel, ok := model.(interface{ Metadata() string }); ok {
		// Render top border with metadata in the center
		//
		// total length of top border runes, not including corners
		length := max(0, width-2-lipgloss.Width(metadataModel.Metadata()))
		leftLength := length / 2
		rightLength := max(0, length-leftLength)
		middle = lipgloss.JoinHorizontal(lipgloss.Left,
			strings.Repeat(border.Top, leftLength),
			metadataModel.Metadata(),
			strings.Repeat(border.Top, rightLength),
		)
	} else {
		middle = strings.Repeat(border.Top, max(0, width-2))
	}
	return lipgloss.NewStyle().
		Foreground(BorderColor(isActive)).
		Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				border.TopLeft,
				middle,
				border.TopRight,
			),
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
