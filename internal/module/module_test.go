package module

import (
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	workdir, _ := internal.NewWorkdir("./testdata/modules")

	t.Run("with_dot_terraform_dir", func(t *testing.T) {
		got := New(workdir, "with_both_s3_backend_and_dot_terraform_dir")

		assert.Equal(t, "with_both_s3_backend_and_dot_terraform_dir", got.Path)
		assert.Nil(t, got.Initialized)
	})

	t.Run("without_dot_terraform_dir", func(t *testing.T) {
		got := New(workdir, "with_s3_backend")
		assert.Equal(t, "with_s3_backend", got.Path)
		if assert.NotNil(t, got.Initialized) {
			assert.Equal(t, false, *got.Initialized)
		}
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
