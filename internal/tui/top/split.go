package top

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/tui"
)

type (
	split struct {
		// The left-hand pane for exploring modules and workspaces
		leftPane tui.ChildModel
		// The right-hand pane for tasks etc
		rightPane tui.ChildModel
		// The currently active pane
		focusedPane tui.ChildModel
		// cached models
		cache *Cache
		// makers for making models for panes
		makers map[tui.Kind]tui.Maker
		// Total width and height of the terminal space available to the split.
		width, height int
		// leftSplitWidth is the width of the left pane when the terminal is split.
		leftSplitWidth int
		// minWidth is the minimum width of each pane.
		minWidth int
	}
)

func newSplit(explorer tui.ChildModel, makers map[tui.Kind]tui.Maker) *split {
	s := &split{
		leftPane:       explorer,
		focusedPane:    explorer,
		cache:          NewCache(),
		makers:         makers,
		minWidth:       20,
		leftSplitWidth: 30,
	}
	s.switchFocusedPane(explorer)
	return s
}

func (s *split) switchFocusedPane(model tui.ChildModel) {
}

func (s *split) Init() tea.Cmd {
	return s.leftPane.Init()
}

func (s *split) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, tui.Keys.Explorer):
			s.switchFocusedPane(s.leftPane)
		case key.Matches(msg, tui.Keys.ShrinkPaneWidth):
			s.updateLeftSplitWidth(-1)
		case key.Matches(msg, tui.Keys.GrowPaneWidth):
			s.updateLeftSplitWidth(1)
		case key.Matches(msg, tui.Keys.SwitchPane):
			if s.focusedPane != s.rightPane {
				s.switchFocusedPane(s.rightPane)
				return nil
			}
			// Right pane is current pane, and it might itself be split into
			// panes, so let it handle the key
			fallthrough
		default:
			return s.focusedPane.Update(msg)
		}
	case tui.FocusExplorerMsg:
		s.switchFocusedPane(s.leftPane)
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.updateLeftSplitWidth(0)
		s.propagateDimensions()
	default:
		return s.focusedPane.Update(msg)
	}
	return nil
}

func (s *split) View() string {
	if s.leftPane == nil || s.rightPane == nil {
		return s.focusedPane.View()
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, s.leftPane.View(), s.rightPane.View())
}

func (s *split) updateLeftSplitWidth(delta int) {
	if s.rightPane == nil {
		// Don't allow user to adjust split width of left pane when there is no split
		return
	}
	s.leftSplitWidth = clamp(s.leftSplitWidth+delta, s.minWidth, s.width-s.minWidth)
	s.propagateDimensions()
}

// propagateDimensions propagates respective dimensions to child models
func (s *split) propagateDimensions() {
	if s.leftPane == nil || s.rightPane == nil {
		_ = s.focusedPane.Update(tea.WindowSizeMsg{
			Width:  s.width,
			Height: s.height,
		})
	} else {
		s.leftPane.Update(tea.WindowSizeMsg{
			Width:  s.leftSplitWidth,
			Height: s.height,
		})
		s.rightPane.Update(tea.WindowSizeMsg{
			Width:  max(s.minWidth, s.width-s.leftSplitWidth),
			Height: s.height,
		})
	}
}

func (p *split) setRightPane(page tui.Page) (cmd tea.Cmd, err error) {
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
	p.rightPane = model
	p.switchFocusedPane(model)
	p.propagateDimensions()
	return cmd, nil
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
