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

	got := New(workdir, "with_s3_backend")
	assert.Equal(t, "with_s3_backend", got.Path)

}

func TestFindModules(t *testing.T) {
	workdir, _ := internal.NewWorkdir("./testdata/modules")
	got, err := findModules(logging.Discard, workdir)
	require.NoError(t, err)

	assert.Equal(t, 3, len(got))
	assert.Contains(t, got, "with_local_backend")
	assert.Contains(t, got, "with_s3_backend")
	assert.Contains(t, got, "with_cloud_backend")
	assert.NotContains(t, got, "broken")
}
