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
	// The test depends upon a .terraform directory being present in testdata.
	// However we ignore .terraform in .gitignore, because it is often created
	// by terraform processes and contains numerous artefacts we don't want in
	// git. Therefore, for this test, we create the directory if it doesn't
	// exist already.
	os.MkdirAll("./testdata/modules/with_both_s3_backend_and_dot_terraform_dir/.terraform", 0o755)

	workdir, _ := internal.NewWorkdir("./testdata/modules")

	t.Run("with_dot_terraform_dir", func(t *testing.T) {
		got := New(workdir, "with_both_s3_backend_and_dot_terraform_dir")

		assert.Equal(t, "with_both_s3_backend_and_dot_terraform_dir", got.Path)
		assert.Nil(t, got.Initialized)
	})

	t.Run("without_dot_terraform_dir", func(t *testing.T) {
		got := New(workdir, "with_s3_backend")
		assert.Equal(t, "with_s3_backend", got.Path)
		assert.Nil(t, got.Initialized)
	})
}

func TestFindModules(t *testing.T) {
	workdir, _ := internal.NewWorkdir("./testdata/modules")
	got, err := findModules(logging.Discard, workdir)
	require.NoError(t, err)

	assert.Equal(t, 4, len(got))
	assert.Contains(t, got, "with_local_backend")
	assert.Contains(t, got, "with_s3_backend")
	assert.Contains(t, got, "with_cloud_backend")
	assert.Contains(t, got, "with_both_s3_backend_and_dot_terraform_dir")
	assert.NotContains(t, got, "broken")
}
