package top

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	"github.com/stretchr/testify/assert"
)

func Test_render(t *testing.T) {
	tests := []struct {
		name     string
		bindings []key.Binding
		want     string
	}{
		{
			"single column",
			[]key.Binding{
				key.NewBinding(key.WithHelp("a", "aaa")),
				key.NewBinding(key.WithHelp("b", "bbb")),
			},
			"a aaa\nb bbb",
		},
		{
			"two columns",
			[]key.Binding{
				key.NewBinding(key.WithHelp("a", "aaa")),
				key.NewBinding(key.WithHelp("b", "bbb")),
				key.NewBinding(key.WithHelp("c", "ccc")),
			},
			"a aaa   c ccc\nb bbb        ",
		},
		{
			"three columns",
			[]key.Binding{
				key.NewBinding(key.WithHelp("a", "aaa")),
				key.NewBinding(key.WithHelp("b", "bbb")),
				key.NewBinding(key.WithHelp("c", "ccc")),
				key.NewBinding(key.WithHelp("d", "ddd")),
				key.NewBinding(key.WithHelp("e", "eee")),
			},
			"a aaa   c ccc   e eee\nb bbb   d ddd        ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortHelpView(tt.bindings, 30)
			assert.Equal(t, tt.want, got)
		})
	}
}
