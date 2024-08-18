package workspace

import (
	"os"
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCost_generateInfracostConfig(t *testing.T) {
	mod := module.New(internal.NewTestWorkdir(t), module.Options{Path: "a/b/c"})
	ws1, err := New(mod, "default")
	require.NoError(t, err)
	ws2, err := New(mod, "dev")
	require.NoError(t, err)

	want := `version: "0.1"
projects:
  - path: a/b/c
    terraform_workspace: default
  - path: a/b/c
    terraform_workspace: dev
`

	got, err := generateCostConfig(ws1, ws2)
	require.NoError(t, err)

	assert.YAMLEq(t, want, string(got))
}

func TestCost_parseBreakdown(t *testing.T) {
	breakdown, err := os.ReadFile("./testdata/costs.json")
	require.NoError(t, err)

	got, err := parseBreakdown(breakdown)
	require.NoError(t, err)

	assert.Len(t, got, 4)
	assert.Contains(t, got, breakdownResult{
		path:      "modules/a",
		workspace: "dev",
		cost:      "239.072",
	})
}
