package app

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest"
)

func TestModules(t *testing.T) {
	wd, _ := os.Getwd()
	t.Log(wd)

	tm := setup(t)

	want := []string{
		"modules/a",
		"modules/b",
		"modules/c",
	}

	teatest.WaitFor(
		t, tm.Output(),
		func(b []byte) bool {
			for _, w := range want {
				if !bytes.Contains(b, []byte(w)) {
					return false
				}
			}
			return true
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)
}
