package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

var (
	shortHelpKeyStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#909090",
		Dark:  "#626262",
	}).Bold(true).Margin(0, 1, 0, 0)

	shortHelpDescStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#B2B2B2",
		Dark:  "#4A4A4A",
	}).Margin(0, 3, 0, 0)
)

// shortHelpView renders help for key bindings across the three rows in the
// main view's header, to the left of the logo.
func shortHelpView(bindings []key.Binding, maxWidth int) string {
	// Create columns of three cells each, one column for keys, and one column
	// for descriptions, repeating until there are no more bindings, or until
	// the maximum width is reached.
	// var cols = make([]string, 2*int(math.Ceil(float64(len(bindings))/3)))

	// enumerate through each group of three bindings, populating a series of
	// pairs of columns, one for keys, one for descriptions
	var (
		pairs []string
		width int
	)
	for i := 0; i < len(bindings); i += 3 {
		var (
			keys  []string
			descs []string
		)
		for j := i; j < min(i+3, len(bindings)); j++ {
			keys = append(keys, bindings[j].Help().Key)
			descs = append(descs, bindings[j].Help().Desc)
		}
		// render a pair of columns, one for keys, one for descs
		pair := lipgloss.JoinHorizontal(lipgloss.Left,
			shortHelpKeyStyle.Render(strings.Join(keys, "\n")),
			shortHelpDescStyle.Render(strings.Join(descs, "\n")),
		)
		// check whether it exceeds the maximum width avail
		width += lipgloss.Width(pair)
		if width > maxWidth {
			break
		}
		pairs = append(pairs, pair)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, pairs...)
}

var (
	longHelpHeadingStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#909090",
		Dark:  "#626262",
	}).Bold(true)
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
