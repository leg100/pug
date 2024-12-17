package tui

import (
	"fmt"
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
	TopLeft BorderPosition = iota
	TopMiddle
	TopRight
	BottomLeft
	BottomMiddle
	BottomRight
)

func borderize(content string, active bool, embeddedText map[BorderPosition]string) string {
	if embeddedText == nil {
		embeddedText = make(map[BorderPosition]string)
	}
	var (
		border = borderThickness[active]
		style  = lipgloss.NewStyle().Foreground(BorderColor(active))
		width  = lipgloss.Width(content)
	)
	encloseInSquareBrackets := func(text string) string {
		if text != "" {
			return fmt.Sprintf("%s%s%s",
				style.Render(border.TopRight),
				text,
				style.Render(border.TopLeft),
			)
		}
		return text
	}
	buildHorizontalBorder := func(leftText, middleText, rightText, leftCorner, inbetween, rightCorner string) string {
		leftText = encloseInSquareBrackets(leftText)
		middleText = encloseInSquareBrackets(middleText)
		rightText = encloseInSquareBrackets(rightText)
		// Construct top border.
		// First determine lengths of each component
		i := width
		i -= lipgloss.Width(leftText)
		i -= lipgloss.Width(middleText)
		i -= lipgloss.Width(rightText)
		// Calculate length of border between embedded texts
		borderLen := max(0, width-lipgloss.Width(leftText)-lipgloss.Width(middleText)-lipgloss.Width(rightText))
		leftBorderLen := borderLen / 2
		rightBorderLen := max(0, i-leftBorderLen)
		// Then construct border string
		s := leftText +
			style.Render(strings.Repeat(inbetween, leftBorderLen)) +
			middleText +
			style.Render(strings.Repeat(inbetween, rightBorderLen)) +
			rightText
		// Make it fit in the space available between the two corners.
		s = lipgloss.NewStyle().
			Inline(true).
			MaxWidth(width).
			Render(s)
		// Add the corners
		return style.Render(leftCorner) + s + style.Render(rightCorner)
	}
	// Stack top border onto remaining borders
	return lipgloss.JoinVertical(lipgloss.Top,
		buildHorizontalBorder(
			embeddedText[TopLeft],
			embeddedText[TopMiddle],
			embeddedText[TopRight],
			border.TopLeft,
			border.Top,
			border.TopRight,
		),
		lipgloss.NewStyle().
			BorderForeground(BorderColor(active)).
			Border(border, false, true, false, true).Render(content),
		buildHorizontalBorder(
			embeddedText[BottomLeft],
			embeddedText[BottomMiddle],
			embeddedText[BottomRight],
			border.BottomLeft,
			border.Bottom,
			border.BottomRight,
		),
	)
}
