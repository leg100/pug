package module

import (
	"os"
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	os.MkdirAll("./testdata/modules/with_both_s3_backend_and_dot_terraform_dir/.terraform", 0o755)

	workdir, _ := internal.NewWorkdir("./testdata/modules")

	got := New(workdir, "with_s3_backend")
	assert.Equal(t, "with_s3_backend", got.Path)

}
