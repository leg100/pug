package module

import (
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
	assert.Contains(t, got, "with_local_backend")
	assert.Contains(t, got, "with_s3_backend")
	assert.Contains(t, got, "with_cloud_backend")
	assert.NotContains(t, got, "broken")
}
