package module

import (
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindModules(t *testing.T) {
	workdir, _ := internal.NewWorkdir("../testdata/configs")
	got, err := findModules(logging.Discard, workdir)
	require.NoError(t, err)

	assert.Equal(t, 7, len(got))
	assert.Contains(t, got, "envs/dev")
	assert.Contains(t, got, "envs/prod")
	assert.Contains(t, got, "envs/staging")
	assert.Contains(t, got, "uninitialized")
	assert.Contains(t, got, "plan_forever")
	assert.Contains(t, got, "multiworkspace")
	assert.Contains(t, got, "with_mods")
}
