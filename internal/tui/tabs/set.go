package tabs

import (
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/navigator"
)

// TabSet is a related set of zero or more tabs, one of which is active, i.e.
// its contents are rendered.
type TabSet struct {
	Tabs []Tab

	// Width and height of the content area
	width  int
	height int

	// The index of the currently active tab
	active int
	// The index of the deafult tab.
	defaultTab int

	info tabSetInfo
}

func NewTabSet(width, height int) TabSet {
	return TabSet{
		width:  width,
		height: height,
	}
}

func (m TabSet) WithTabSetInfo(i tabSetInfo) TabSet {
	m.info = i
	return m
}

// Init initializes the existing tabs in the collection.
func (m TabSet) Init() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.Tabs))
	for i, tab := range m.Tabs {
		cmds[i] = tab.Init()
	}
	return tea.Batch(cmds...)
}

var ErrDuplicateTab = errors.New("not allowed to create tabs with duplicate titles")

// AddTab adds a tab to the tab set, using the maker and parent to construct the
// model associated with the tab. The title must be unique in the set. Upon
// success the associated model's Init() is returned for the caller to
// initialise the model.
func (m *TabSet) AddTab(maker tui.Maker, parent resource.Resource, title tui.TabTitle) (tea.Cmd, error) {
	for _, tab := range m.Tabs {
		if tab.Title == title {
			return nil, ErrDuplicateTab
		}
	}

	model, err := maker.Make(parent, m.contentWidth(), m.contentHeight())
	if err != nil {
		return nil, err
	}
	m.Tabs = append(m.Tabs, Tab{Model: model, Title: title})
	return model.Init(), nil
}

// Active returns the currently active tab. If there are no tabs, then false is
// returned.
func (m TabSet) Active() (bool, Tab) {
	if len(m.Tabs) > 0 {
		return true, m.Tabs[m.active]
	}
	return false, Tab{}
}

// Active returns the title of the currently active tab. If there are no tabs,
// then an empty string is returned.
func (m TabSet) ActiveTitle() tui.TabTitle {
	if len(m.Tabs) > 0 {
		return m.Tabs[m.active].Title
	}
	return ""
}

// setActiveTab makes the tab with the given title the active tab. If title is an
// empty string, then the first tab is made the active tab. If title is
// non-empty and there is no tab with a matching title then no action is taken.
func (m *TabSet) setActiveTab(title tui.TabTitle) {
	if title == "" && len(m.Tabs) > 0 {
		m.active = m.defaultTab
		return
	}

	for i, tab := range m.Tabs {
		if tab.Title == title {
			m.active = i
		}
	}
}

// SetDefaultTab sets the default tab and immediately makes it the active tab.
// If the tab set receives an empty SetActiveTabMsg then this is the tab that is
// by default made the active tab. If no tab exists with the given title then no
// actiion is taken.
func (m *TabSet) SetDefaultTab(title tui.TabTitle) {
	for i, tab := range m.Tabs {
		if tab.Title == title {
			m.defaultTab = i
			m.active = i
		}
	}
}

func (m *TabSet) HelpBindings() (bindings []key.Binding) {
	if len(m.Tabs) > 0 {
		active := m.Tabs[m.active].Model
		if bindings, ok := active.(tui.ModelHelpBindings); ok {
			return bindings.HelpBindings()
		}
	}
	return nil
}

func (m *TabSet) getTab(tabIndex int) (Tab, bool) {
	if len(m.Tabs) == 0 {
		return Tab{}, false
	}

	if tabIndex < 0 {
		// If negative index then retrieve last tab
		tabIndex = len(m.Tabs) - 1
	} else if tabIndex > len(m.Tabs)-1 {
		// If beyond bounds then retrieve first tab
		tabIndex = 0
	}

	return m.Tabs[tabIndex], true
}

func (m TabSet) Update(msg tea.Msg) (TabSet, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		var cmd tea.Cmd
		switch {
		case key.Matches(msg, keys.Navigation.TabNext):
			// Cycle tabs, going back to the first tab after the last tab
			if tab, ok := m.getTab(m.active + 1); ok {
				cmd = navigator.SwitchTab(tab.Title)
			}
		case key.Matches(msg, keys.Navigation.TabLast):
			// Cycle back thru tabs, going to the last tab after the first tab.
			if tab, ok := m.getTab(m.active - 1); ok {
				cmd = navigator.SwitchTab(tab.Title)
			}
		default:
			// Send other keys to active tab
			cmd = m.updateActive(msg)
		}
		return m, cmd
	case tui.SetActiveTabMsg:
		m.setActiveTab(tui.TabTitle(msg))
	case tui.FilterFocusReqMsg, tui.FilterCloseMsg, tui.FilterKeyMsg:
		// Send filter messages to the active tab only.
		cmd := m.updateActive(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Relay modified resize message onto each tab model
		m.updateTabs(tea.WindowSizeMsg{
			Width:  m.contentWidth(),
			Height: m.contentHeight(),
		})
		return m, nil
	}

	// Updates each tab's respective model in-place.
	cmds = append(cmds, m.updateTabs(msg))

	return m, tea.Batch(cmds...)
}

func (m *TabSet) updateTabs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.Tabs))
	for i := range m.Tabs {
		cmds[i] = m.updateTab(i, msg)
	}
	return tea.Batch(cmds...)
}

func (m *TabSet) updateTab(tabIndex int, msg tea.Msg) tea.Cmd {
	updated, cmd := m.Tabs[tabIndex].Update(msg)
	m.Tabs[tabIndex].Model = updated
	return cmd
}

func (m *TabSet) updateActive(msg tea.Msg) tea.Cmd {
	if ok, _ := m.Active(); ok {
		cmd := m.updateTab(m.active, msg)
		return cmd
	}
	return nil
}

var (
	activeTabStyle   = tui.Bold.Copy().Foreground(tui.ActiveTabColor)
	inactiveTabStyle = tui.Regular.Copy().Foreground(tui.InactiveTabColor)
)

func (m TabSet) View() string {
	var (
		tabHeaders       []string
		tabsHeadersWidth int
	)
	for i, t := range m.Tabs {
		var (
			headingStyle  lipgloss.Style
			underlineChar string
		)
		if i == m.active {
			headingStyle = activeTabStyle.Copy()
			underlineChar = "━"
		} else {
			headingStyle = inactiveTabStyle.Copy()
			underlineChar = "─"
		}
		heading := headingStyle.Copy().Padding(0, 1).Render(string(t.Title))
		if status, ok := t.Model.(tabStatus); ok {
			heading += headingStyle.Copy().Bold(false).Padding(0, 1, 0, 0).Render(status.TabStatus())
		}
		underline := headingStyle.Render(strings.Repeat(underlineChar, tui.Width(heading)))
		rendered := lipgloss.JoinVertical(lipgloss.Top, heading, underline)
		tabHeaders = append(tabHeaders, rendered)
		tabsHeadersWidth += tui.Width(heading)
	}

	// Populate remaining space to the right of the tab headers with a faint
	// grey underline. If the tab set parent implements tabSetInfo then that'll
	// be called and its contents rendered above the underline.
	remainingWidth := max(0, m.width-tabsHeadersWidth)
	var rightSideInfo string
	if m.info != nil {
		rightSideInfo = tui.Padded.Copy().
			Width(remainingWidth).
			Align(lipgloss.Right).
			Render(m.info.TabSetInfo(m.ActiveTitle()))
	}
	tabHeadersFiller := lipgloss.JoinVertical(lipgloss.Top,
		rightSideInfo,
		inactiveTabStyle.Copy().Render(strings.Repeat("─", remainingWidth)),
	)
	tabHeaders = append(tabHeaders, tabHeadersFiller)

	// Join tab headers and filler together
	tabHeadersContainer := lipgloss.JoinHorizontal(lipgloss.Bottom, tabHeaders...)

	var tabContent string
	if len(m.Tabs) > 0 {
		tabContent = m.Tabs[m.active].View()
	}
	return lipgloss.JoinVertical(lipgloss.Top, tabHeadersContainer, tabContent)
}

// Width of the tab content area
func (m TabSet) contentWidth() int {
	return m.width
}

// Height of the tab content area
func (m TabSet) contentHeight() int {
	return m.height - tabHeaderHeight
}
