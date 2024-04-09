package run

import (
	"fmt"
	"testing"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/testutils"
	"github.com/leg100/pug/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_MakePugDirectory(t *testing.T) {
	mod := module.New("a/b/c")
	ws, err := workspace.New(mod, "dev")
	require.NoError(t, err)

	testutils.ChTempDir(t)

	run, err := newRun(mod, ws, CreateOptions{})
	require.NoError(t, err)

	want := fmt.Sprintf("a/b/c/.pug/dev/%s", run.ID)
	assert.DirExists(t, want)
}

func TestRun_PugDirectory(t *testing.T) {
	mod := module.New("a/b/c")
	ws, err := workspace.New(mod, "dev")
	require.NoError(t, err)

	run, err := newRun(mod, ws, CreateOptions{})
	require.NoError(t, err)

	want := fmt.Sprintf(".pug/dev/%s", run.ID)
	assert.Equal(t, want, run.artefactsPath)
}

func TestRun_PlanPath(t *testing.T) {
	mod := module.New("a/b/c")
	ws, err := workspace.New(mod, "dev")
	require.NoError(t, err)

	run, err := newRun(mod, ws, CreateOptions{})
	require.NoError(t, err)

	want := fmt.Sprintf(".pug/dev/%s/plan.out", run.ID)
	assert.Equal(t, want, run.PlanPath())
}
