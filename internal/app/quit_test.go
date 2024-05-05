package app

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

func TestQuit(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	tm.Send(tea.KeyMsg{
		Type: tea.KeyCtrlC,
	})

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Quit pug (y/N)? ")
	})

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'y'},
	})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
