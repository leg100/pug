package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// pane models expose metadata to be embedded in certain positions in borders:
// options:
// * one method per position (Title(), TopLeft(), etc)
// * one method returning positions: Metadata() map[border-position]string

// inputs:
// * content to wrap with border, from which width and height is determined
// * metadata to place within certain positions in border:
// map[border-position]string; or use functional options,
// WithBottomLeft(string), etc.
// * border style; alternatively let method define style based on a
// border-active input
// output:
// * content wrapped with border.

// positions:
// * Title (table row info)
// * TopLeft (task info; resource info; log msg info; task group info)
// * BottomLeft (task status; task summary; state summary)
// * BottomRight (scroll percentage)
//

var borderThickness = map[bool]lipgloss.Border{
	true:  lipgloss.Border(lipgloss.ThickBorder()),
	false: lipgloss.Border(lipgloss.NormalBorder()),
}

type BorderPosition int

const (
	TopMiddle BorderPosition = iota
	TopLeft
	BottomLeft
	BottomMiddle
)

func borderize(content string, active bool, embeddedText map[BorderPosition]string) string {
	if embeddedText == nil {
		embeddedText = make(map[BorderPosition]string)
	}
	var (
		border = borderThickness[active]
		style  = lipgloss.NewStyle().Foreground(BorderColor(active))
	)
	buildHorizontalBorder := func(leftText, middleText, leftCorner, inbetween, rightCorner string) string {
		// Construct top border.
		// First determine lengths of each component
		i := lipgloss.Width(content)
		i -= lipgloss.Width(leftText)
		i -= lipgloss.Width(middleText)
		rightBorderLen := min((lipgloss.Width(content)-lipgloss.Width(middleText))/2, i)
		rightBorderLen = max(rightBorderLen, 0)
		leftBorderLen := max(0, i-rightBorderLen)
		// Then construct border string
		s := leftText +
			style.Render(strings.Repeat(inbetween, leftBorderLen)) +
			middleText +
			style.Render(strings.Repeat(inbetween, rightBorderLen))
		// Make it fit in the space available between the two corners.
		s = lipgloss.NewStyle().
			Inline(true).
			MaxWidth(lipgloss.Width(content)).
			Render(s)
		// Add the corners
		return style.Render(leftCorner) + s + style.Render(rightCorner)
	}
	// Stack top border onto remaining borders
	return lipgloss.JoinVertical(lipgloss.Top,
		buildHorizontalBorder(embeddedText[TopLeft], embeddedText[TopMiddle], border.TopLeft, border.Top, border.TopRight),
		lipgloss.NewStyle().
			BorderForeground(BorderColor(active)).
			Border(border, false, true, false, true).Render(content),
		buildHorizontalBorder(embeddedText[BottomLeft], embeddedText[BottomMiddle], border.BottomLeft, border.Bottom, border.BottomRight),
	)
}
