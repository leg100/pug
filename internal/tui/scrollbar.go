package tui

import (
	"math"
	"strings"
)

const (
	ScrollbarWidth = 1

	scrollbarThumb = "█"
	scrollbarTrack = "░"
)

func Scrollbar(height, total, visible, offset int) string {
	ratio := float64(height) / float64(total)
	thumbHeight := max(1, int(math.Round(float64(visible)*ratio)))
	thumbOffset := max(0, min(height-thumbHeight, int(math.Round(float64(offset)*ratio))))

	return strings.TrimRight(
		strings.Repeat(scrollbarTrack+"\n", thumbOffset)+
			strings.Repeat(scrollbarThumb+"\n", thumbHeight)+
			strings.Repeat(scrollbarTrack+"\n", max(0, height-thumbOffset-thumbHeight)),
		"\n",
	)
}
