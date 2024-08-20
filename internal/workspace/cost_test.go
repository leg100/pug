package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCost_generateInfracostConfig(t *testing.T) {
	workdir := internal.NewTestWorkdir(t)
	mod := module.New(module.Options{Path: "a/b/c"})
	ws1, err := New(mod, "default")
	require.NoError(t, err)
	ws2, err := New(mod, "dev")
	require.NoError(t, err)

	// Create a workspace tfvars file for ws2
	path := workdir.Join(mod.Path, "dev.tfvars")
	os.MkdirAll(filepath.Dir(path), 0o755)
	_, err = os.Create(path)
	require.NoError(t, err)

	want := `version: "0.1"
projects:
  - path: a/b/c
    terraform_workspace: default
  - path: a/b/c
    terraform_workspace: dev
    terraform_var_files:
      - dev.tfvars
`

	got, err := generateCostConfig(workdir, ws1, ws2)
	require.NoError(t, err)

	assert.YAMLEq(t, want, string(got))
}

func TestCost_parseBreakdown(t *testing.T) {
	breakdown, err := os.ReadFile("./testdata/costs.json")
	require.NoError(t, err)

	got, err := parseBreakdown(breakdown)
	require.NoError(t, err)

	assert.Equal(t, "264.248", got.total)
	assert.Len(t, got.projects, 4)
	assert.Contains(t, got.projects, breakdownResultProject{
		path:      "modules/a",
		workspace: "dev",
		cost:      "239.072",
	})
}
