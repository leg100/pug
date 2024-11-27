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

type Position int

const (
	TopRightPane Position = iota
	BottomRightPane
	LeftPane
)

// PaneManager manages the layout of the three panes that compose the Pug full screen terminal app.
type PaneManager struct {
	// makers for making models for panes
	makers map[Kind]Maker
	// cached models
	cache *Cache
	// the position of the currently active pane
	active Position
	// panes tracks currently visible panes
	panes map[Position]*pane
	// total width and height of the terminal space available to panes.
	width, height int
	// minimum width and heights for panes
	minWidth, minHeight int
	// maximum width of left pane and maximum height of top right pane
	maxLeftPaneWidth, maxTopRightHeight int
}

type pane struct {
	page   Page
	width  int
	height int
}

// NewPaneManager constructs the pane manager with at least the explorer, which
// occupies the left pane.
func NewPaneManager(explorer tea.Model, makers map[Kind]Maker) *PaneManager {
	cache := NewCache()
	cache.Put(Page{Kind: ExplorerKind}, explorer)

	return &PaneManager{
		makers: makers,
		cache:  cache,
		active: LeftPane,
		panes: map[Position]*pane{
			LeftPane: {
				page: Page{Kind: ExplorerKind},
			},
		},
		minWidth:          20,
		minHeight:         20,
		maxLeftPaneWidth:  20,
		maxTopRightHeight: 20,
	}
}

func (p *PaneManager) Init() tea.Cmd {
	return p.getModel(p.active).Init()
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
			p.changeActivePaneWidth(-1)
		case key.Matches(msg, Keys.GrowPaneWidth):
			p.changeActivePaneWidth(1)
		case key.Matches(msg, Keys.ShrinkPaneHeight):
			p.changeActivePaneHeight(-1)
		case key.Matches(msg, Keys.GrowPaneHeight):
			p.changeActivePaneHeight(1)
		case key.Matches(msg, Keys.SwitchPane):
			p.switchActivePane()
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
		p.setPaneWidths()
		p.setPaneHeights()
		p.updateChildSizes()
	default:
		// Send remaining msg types to all cached models
		cmds = append(cmds, p.cache.UpdateAll(msg)...)
	}
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

func (p *PaneManager) SetPane(page Page, position Position) (cmd tea.Cmd, err error) {
	model := p.cache.Get(page)
	if model == nil {
		maker, ok := p.makers[page.Kind]
		if !ok {
			return nil, fmt.Errorf("no maker could be found for %s", page.Kind)
		}
		model, err = maker.Make(page.ID, 0, 0)
		if err != nil {
			return nil, fmt.Errorf("making page: %w", err)
		}
		p.cache.Put(page, model)
		cmd = model.Init()
	}
	p.panes[position] = &pane{page: page}
	p.active = position
	p.setPaneWidths()
	p.setPaneHeights()
	p.updateChildSizes()
	return cmd, nil
}

// ActiveModel retrieves the model of the active pane.
func (p *PaneManager) ActiveModel() tea.Model {
	return p.getModel(p.active)
}

func (p *PaneManager) switchActivePane() {
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
	p.switchActivePane()

	p.setPaneHeights()
	p.setPaneWidths()
	p.updateChildSizes()
	return nil
}

func (p *PaneManager) setPaneWidths() {
	if p.panes[LeftPane] != nil {
		if p.panes[TopRightPane] != nil || p.panes[BottomRightPane] != nil {
			p.panes[LeftPane].width = clamp(p.panes[LeftPane].width, p.minWidth, p.width-p.minWidth)
		} else {
			p.panes[LeftPane].width = p.width
		}
	}
	if p.panes[TopRightPane] != nil {
		if p.panes[LeftPane] != nil {
			p.panes[TopRightPane].width = max(p.minWidth, p.width-p.panes[LeftPane].width)
		} else {
			p.panes[TopRightPane].width = p.width
		}
	}
	if p.panes[BottomRightPane] != nil {
		if p.panes[LeftPane] != nil {
			p.panes[BottomRightPane].width = max(p.minWidth, p.width-p.panes[LeftPane].width)
		} else {
			p.panes[BottomRightPane].width = p.width
		}
	}
}

func (p *PaneManager) setPaneHeights() {
	if p.panes[LeftPane] != nil {
		p.panes[LeftPane].height = p.height
	}
	if p.panes[TopRightPane] != nil {
		if p.panes[BottomRightPane] != nil {
			p.panes[TopRightPane].height = clamp(p.panes[TopRightPane].height, p.minHeight, p.height-p.minHeight)
		} else {
			p.panes[TopRightPane].height = p.height
		}
	}
	if p.panes[BottomRightPane] != nil {
		if p.panes[TopRightPane] != nil {
			p.panes[BottomRightPane].height = max(p.minHeight, p.height-p.panes[TopRightPane].height)
		} else {
			p.panes[BottomRightPane].height = p.height
		}
	}
}

func (p *PaneManager) changeActivePaneWidth(delta int) {
	switch p.active {
	case TopRightPane, BottomRightPane:
		// on the right panes, shrink width is actually grow width, and vice
		// versa
		delta = -delta
	}
	for position := range p.panes {
		if position == p.active {
			p.panes[position].width = clamp(p.panes[position].width+delta, p.minWidth, p.width-p.minWidth)
		} else {
			p.panes[position].width = clamp(p.panes[position].width-delta, p.minWidth, p.width-p.minWidth)
		}
	}
	p.maxLeftPaneWidth = p.panes[LeftPane].width
	p.setPaneWidths()
	p.updateChildSizes()
}

func (p *PaneManager) changeActivePaneHeight(delta int) {
	if p.active == LeftPane {
		// Cannot change height of left pane because it occupies the full height
		// already.
		return
	}
	for position := range p.panes {
		if position == p.active {
			p.panes[position].height = clamp(p.panes[position].height+delta, p.minHeight, p.height-p.minHeight)
		} else {
			p.panes[position].height = clamp(p.panes[position].height-delta, p.minHeight, p.height-p.minHeight)
		}
	}
	p.maxTopRightHeight = p.panes[TopRightPane].height
	p.setPaneHeights()
	p.updateChildSizes()
}

func (p *PaneManager) updateChildSizes() {
	for position, pane := range p.panes {
		p.updateModel(position, tea.WindowSizeMsg{
			Width:  pane.width - 2,
			Height: pane.height - 2,
		})
	}
}

func (p *PaneManager) getModel(position Position) tea.Model {
	pane := p.panes[position]
	return p.cache.Get(pane.page)
}

func (p *PaneManager) updateModel(position Position, msg tea.Msg) tea.Cmd {
	pane := p.panes[position]
	return p.cache.Update(pane.page, msg)
}

var border = map[bool]lipgloss.Border{
	true:  lipgloss.Border(lipgloss.ThickBorder()),
	false: lipgloss.Border(lipgloss.NormalBorder()),
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
	if m.panes[position] == nil {
		return ""
	}
	model := m.getModel(position)
	renderedPane := model.View()
	isActive := position == m.active
	border := border[isActive]
	var topBorder string
	if metadataModel, ok := model.(interface{ Metadata() string }); ok {
		// Render top border with metadata in the center
		//
		// total length of top border runes, not including corners
		length := max(0, m.panes[position].width-2-lipgloss.Width(metadataModel.Metadata()))
		leftLength := length / 2
		rightLength := max(0, length-leftLength)
		topBorder = lipgloss.JoinHorizontal(lipgloss.Left,
			border.TopLeft,
			strings.Repeat(border.Top, leftLength),
			metadataModel.Metadata(),
			strings.Repeat(border.Top, rightLength),
			border.TopRight,
		)
	} else {
		topBorder = border.TopLeft + strings.Repeat(border.Top, max(0, m.panes[position].width-2)) + border.TopRight
	}
	// TODO: border color
	return lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().Render(topBorder),
		lipgloss.NewStyle().Border(border, false, true, true, true).Render(renderedPane),
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
