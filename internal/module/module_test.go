package module

import (
	"os"
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	os.MkdirAll("./testdata/modules/with_both_s3_backend_and_dot_terraform_dir/.terraform", 0o755)

	workdir, _ := internal.NewWorkdir("./testdata/modules")

	got := New(workdir, Options{Path: "with_s3_backend", Backend: "s3"})
	assert.Equal(t, "with_s3_backend", got.Path)

}

func TestFindModules(t *testing.T) {
	workdir, _ := internal.NewWorkdir("./testdata/modules")
	got, err := findModules(logging.Discard, workdir)
	require.NoError(t, err)

	assert.Equal(t, 4, len(got), got)
	assert.Contains(t, got, Options{Path: "with_local_backend", Backend: "local"})
	assert.Contains(t, got, Options{Path: "with_s3_backend", Backend: "s3"})
	assert.Contains(t, got, Options{Path: "with_cloud_backend", Backend: "cloud"})
	assert.Contains(t, got, Options{Path: "terragrunt_with_local", Backend: "local"})
	assert.NotContains(t, got, "broken")
}
