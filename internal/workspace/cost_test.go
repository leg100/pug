package workspace

import (
	"os"
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOverallCost(t *testing.T) {
	in, err := os.Open("./testdata/infracost.table")
	require.NoError(t, err)

	got, err := parseInfracostOutput(in)
	require.NoError(t, err)

	assert.Equal(t, "$9.27", got)
}

func TestCost_generateInfracostConfig(t *testing.T) {
	mod := module.New(internal.NewTestWorkdir(t), module.Options{Path: "a/b/c"})
	ws, err := New(mod, "dev")
	require.NoError(t, err)

	want := `version: "0.1"
projects:
  - path: a/b/c
  - path: a/b/c
    terraform_workspace: dev
`

	got, err := generateInfracostConfig(mod, ws)
	require.NoError(t, err)

	assert.YAMLEq(t, want, string(got))
}
