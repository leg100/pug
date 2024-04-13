package app

import (
	"context"
	"testing"

	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) *teatest.TestModel {
	t.Helper()

	// Cancel context once test finishes.
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	app, m, err := newApp(
		config{
			FirstPage: "modules",
			Program:   "terraform",
			Workdir:   "./testdata",
		},
	)
	require.NoError(t, err)

	tm := teatest.NewTestModel(
		t,
		m,
		teatest.WithInitialTermSize(300, 100),
	)
	app.start(ctx, tm)
	return tm

}
