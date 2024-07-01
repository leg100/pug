package module

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindModules(t *testing.T) {
	workdir, _ := internal.NewWorkdir("./testdata/modules")
	got, err := findModules(logging.Discard, workdir)
	require.NoError(t, err)

	assert.Equal(t, 3, len(got))
	assert.Contains(t, got, findResult{path: "with_local_backend"})
	assert.Contains(t, got, findResult{path: "with_s3_backend"})
	assert.Contains(t, got, findResult{path: "with_cloud_backend"})
	assert.NotContains(t, got, "broken")
}

func TestFindTerragruntModules(t *testing.T) {
	buf := new(bytes.Buffer)

	cmd := exec.Command("terragrunt", "graph-dependencies")
	cmd.Dir = "./testdata/terragrunt"
	cmd.Stdout = buf
	err := cmd.Run()
	require.NoError(t, err)

	got, err := findTerragruntModules(buf)
	require.NoError(t, err)

	// Should find 4 modules
	assert.Len(t, got, 4)

	hasFindResult(t, got, findResult{
		path: "root/backend-app",
		dependencies: []string{
			"root/mysql",
			"root/redis",
			"root/vpc",
		},
	})
	hasFindResult(t, got, findResult{
		path: "root/frontend-app",
		dependencies: []string{
			"root/backend-app",
			"root/vpc",
		},
	})
	hasFindResult(t, got, findResult{
		path: "root/mysql",
		dependencies: []string{
			"root/vpc",
		},
	})
	hasFindResult(t, got, findResult{
		path: "root/redis",
		dependencies: []string{
			"root/vpc",
		},
	})
	hasFindResult(t, got, findResult{
		path: "root/vpc",
	})
}

func hasFindResult(t *testing.T, gotResults []findResult, want findResult) {
	t.Helper()

	for _, got := range gotResults {
		if want.path == got.path {
			if assert.Equal(t, len(want.dependencies), len(got.dependencies)) {
				for _, wantDep := range want.dependencies {
					assert.Contains(t, got.dependencies, wantDep)
				}
				return
			}
		}
	}
	t.Errorf("failed to find %v in %v", want, gotResults)
}
