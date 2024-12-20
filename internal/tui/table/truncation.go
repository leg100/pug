package table

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var defaultTruncationFunc = TruncateRight

type TruncationFunc func(s string, w int, tailOrPrefix string) string

func TruncateRight(s string, w int, tail string) string {
	return ansi.Truncate(s, w, tail)
}

func TruncateLeft(s string, w int, prefix string) string {
	// ansi.TruncateLeft is weird and doesn't obey its documented behaviour:
	// instead it removes n chars from the left-side of the string and prefixes
	// the prefix string. Because it is ANSI aware it is still useful. It is
	// only called if the string is determined to need truncating.
	if overlap := lipgloss.Width(s) - w; overlap > 0 {
		// needs truncating - truncate off the overlapping number of chars as
		// well as the width of the prefix so that once the prefix is prepended
		// the length obeys w.
		return ansi.TruncateLeft(s, overlap+lipgloss.Width(prefix), prefix)
	}
	return s
}
