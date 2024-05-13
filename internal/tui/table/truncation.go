package table

import (
	"github.com/leg100/go-runewidth"
	"github.com/leg100/reflow/truncate"
)

var defaultTruncationFunc = TruncateRight

type TruncationFunc func(s string, w int, tailOrPrefix string) string

func TruncateRight(s string, w int, tail string) string {
	return truncate.StringWithTail(s, uint(w), tail)
}

func TruncateLeft(s string, w int, prefix string) string {
	return runewidth.TruncateLeft(s, w, prefix)
}
