package top

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/leg100/pug/internal/logging"
	"github.com/stretchr/testify/require"
)

func TestQuit(t *testing.T) {
	m, err := New(Options{
		FirstPage:        "modules",
		Logger:           &logging.Logger{},
		Debug:            true,
		ModuleService:    &fakeModuleService{},
		WorkspaceService: &fakeWorkspaceService{},
		TaskService:      &fakeTaskService{},
	})
	require.NoError(t, err)

	tm := teatest.NewTestModel(
		t,
		m,
		teatest.WithInitialTermSize(300, 100),
	)

	tm.Send(tea.KeyMsg{
		Type: tea.KeyCtrlC,
	})

	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("Quit pug? (y/N): "))
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
