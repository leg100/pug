package tui

import (
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ScrollbarWidth is the width of the scrollbar. Hardcoded for performance
// reasons.
const ScrollbarWidth = 2

// NewScrollbar create a new vertical scrollbar.
func NewScrollbar() *Scrollbar {
	return &Scrollbar{
		Style:      lipgloss.NewStyle().Width(2),
		ThumbStyle: lipgloss.NewStyle().SetString("█"),
		TrackStyle: lipgloss.NewStyle().SetString("░"),
	}
}

// Scrollbar is a model for a vertical scrollbar
type Scrollbar struct {
	Style       lipgloss.Style
	ThumbStyle  lipgloss.Style
	TrackStyle  lipgloss.Style
	height      int
	thumbHeight int
	thumbOffset int
}

// Init initializes the scrollbar model.
func (m Scrollbar) Init() tea.Cmd {
	return nil
}

func (m *Scrollbar) SetHeight(height int) {
	m.height = height
}

func (m *Scrollbar) ComputeThumb(total, visible, offset int) {
	ratio := float64(m.height) / float64(total)

	m.thumbHeight = max(1, int(math.Round(float64(visible)*ratio)))
	m.thumbOffset = max(0, min(m.height-m.thumbHeight, int(math.Round(float64(offset)*ratio))))
}

// View renders the scrollbar to a string.
func (m Scrollbar) View() string {
	bar := strings.TrimRight(
		strings.Repeat(m.TrackStyle.String()+"\n", m.thumbOffset)+
			strings.Repeat(m.ThumbStyle.String()+"\n", m.thumbHeight)+
			strings.Repeat(m.TrackStyle.String()+"\n", max(0, m.height-m.thumbOffset-m.thumbHeight)),
		"\n",
	)
	return m.Style.Align(lipgloss.Right).Render(bar)
}
