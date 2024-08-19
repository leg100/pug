package module

import (
	"context"
	"os"
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	os.MkdirAll("./testdata/modules/with_both_s3_backend_and_dot_terraform_dir/.terraform", 0o755)

	factory := &factory{&fakeWorkspaceLoader{}}

	got, err := factory.newModule(Options{Path: "with_s3_backend", Backend: "s3"})
	require.NoError(t, err)
	assert.Equal(t, "with_s3_backend", got.Path)
}

func TestFindModules(t *testing.T) {
	workdir, _ := internal.NewWorkdir("./testdata/modules")
	modules, errch := find(context.Background(), workdir)

	var got []Options
	for opts := range modules {
		got = append(got, opts)
	}

	assert.Equal(t, 6, len(got), got)
	assert.Contains(t, got, Options{Path: "with_local_backend", Backend: "local"})
	assert.Contains(t, got, Options{Path: "with_s3_backend", Backend: "s3"})
	assert.Contains(t, got, Options{Path: "with_cloud_backend", Backend: "cloud"})
	assert.Contains(t, got, Options{Path: "terragrunt_with_local", Backend: "local"})
	assert.Contains(t, got, Options{Path: "terragrunt_without_backend", Backend: ""})
	assert.Contains(t, got, Options{Path: "multiple_tf_files", Backend: "local"})
	assert.NotContains(t, got, "broken")

	// Expect one error from broken module then error channel should close
	goterr := <-errch
	assert.Contains(t, goterr.Error(), "Unclosed configuration block")
	_, closed := <-errch
	assert.False(t, closed)
}
