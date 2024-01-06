package keys

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/stretchr/testify/assert"
)

func Test_keyMapToSlice(t *testing.T) {
	got := KeyMapToSlice(viewport.DefaultKeyMap())
	want := []key.Binding(
		[]key.Binding{
			key.NewBinding(
				key.WithKeys("pgdown", " ", "f"),
				key.WithHelp("f/pgdn", "page down"),
			),
			key.NewBinding(
				key.WithKeys("pgup", "b"),
				key.WithHelp("b/pgup", "page up"),
			),
			key.NewBinding(
				key.WithKeys("u", "ctrl+u"),
				key.WithHelp("u", "½ page up"),
			),
			key.NewBinding(
				key.WithKeys("d", "ctrl+d"),
				key.WithHelp("d", "½ page down"),
			),
			key.NewBinding(
				key.WithKeys("down", "j"),
				key.WithHelp("↓/j", "down"),
			),
			key.NewBinding(
				key.WithKeys("up", "k"),
				key.WithHelp("↑/k", "up"),
			),
		})
	assert.Equal(t, want, got)
}
