package top

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/tui"
)

var (
	shortHelpKeyStyle  = tui.Bold.Foreground(tui.HelpKey).Copy().Margin(0, 1, 0, 0)
	shortHelpDescStyle = tui.Regular.Foreground(tui.HelpDesc).Copy()
)

// shortHelpView renders help for key bindings within the header.
func shortHelpView(bindings []key.Binding, maxWidth int) string {
	// enumerate through each group of three bindings, populating a series of
	// pairs of columns, one for keys, one for descriptions
	var (
		b     strings.Builder
		width int
	)
	for i, binding := range bindings {
		var bb strings.Builder
		if i > 0 {
			bb.WriteString("  ")
		}
		bb.WriteString(shortHelpKeyStyle.Render(binding.Help().Key))
		bb.WriteString(" ")
		bb.WriteString(shortHelpDescStyle.Render(binding.Help().Desc))

		// check whether it exceeds the maximum width avail
		width += lipgloss.Width(bb.String())
		if width > maxWidth {
			break
		}
		b.WriteString(bb.String())
	}
	return b.String()
}

var (
	longHelpHeadingStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#909090",
		Dark:  "#626262",
	}).Bold(true).Margin(0, 3, 0, 0)

	longHelpKeyStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#909090",
		Dark:  "#626262",
	}).Bold(true).Margin(0, 1, 0, 0)

	longHelpDescStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#B2B2B2",
		Dark:  "#4A4A4A",
	}).Margin(0, 3, 0, 0)
)

// fullHelpView renders a table of three columns describing the key bindings,
// categorised into resource, general, and navigation keys.
func fullHelpView(resource, general, navigation []key.Binding) string {
	resourceKeys := make([]string, len(resource))
	resourceDescs := make([]string, len(resource))
	for i, kb := range resource {
		resourceKeys[i] = longHelpKeyStyle.Render(kb.Help().Key)
		resourceDescs[i] = longHelpDescStyle.Render(kb.Help().Desc)
	}

	generalKeys := make([]string, len(general))
	generalDescs := make([]string, len(general))
	for i, kb := range general {
		generalKeys[i] = longHelpKeyStyle.Render(kb.Help().Key)
		generalDescs[i] = longHelpDescStyle.Render(kb.Help().Desc)
	}

	navigationKeys := make([]string, len(navigation))
	navigationDescs := make([]string, len(navigation))
	for i, kb := range navigation {
		navigationKeys[i] = longHelpKeyStyle.Render(kb.Help().Key)
		navigationDescs[i] = longHelpDescStyle.Render(kb.Help().Desc)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.JoinVertical(lipgloss.Top,
			longHelpHeadingStyle.Render("RESOURCE"),
			lipgloss.JoinHorizontal(lipgloss.Left,
				strings.Join(resourceKeys, "\n"),
				strings.Join(resourceDescs, "\n"),
			),
		),
		lipgloss.JoinVertical(lipgloss.Top,
			longHelpHeadingStyle.Render("GENERAL"),
			lipgloss.JoinHorizontal(lipgloss.Left,
				strings.Join(generalKeys, "\n"),
				strings.Join(generalDescs, "\n"),
			),
		),
		lipgloss.JoinVertical(lipgloss.Top,
			longHelpHeadingStyle.Render("NAVIGATION"),
			lipgloss.JoinHorizontal(lipgloss.Left,
				strings.Join(navigationKeys, "\n"),
				strings.Join(navigationDescs, "\n"),
			),
		),
	)
}
