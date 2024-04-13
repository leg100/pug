package app

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

func TestQuit(t *testing.T) {
	tm := setup(t)

	tm.Send(tea.KeyMsg{
		Type: tea.KeyCtrlC,
	})

	teatest.WaitFor(
		t, tm.Output(),
		func(b []byte) bool {
			return bytes.Contains(b, []byte("Quit pug (y/N)? "))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'y'},
	})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
