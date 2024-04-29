package app

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func init() {
	// Disable color in tests (see https://charm.sh/blog/teatest/)
	lipgloss.SetColorProfile(termenv.Ascii)
}
