package workspace

import (
	"testing"

	"github.com/leg100/pug/internal/module"
	"github.com/stretchr/testify/assert"
)

// TODO: rewrite as test for resetWorkspaces()
//
// func TestFindWorkspaces(t *testing.T) {
// 	got, err := findWorkspaces([]Module{
// 		{"testdata/configs/envs/dev", true},
// 		{"testdata/configs/envs/prod", true},
// 		{"testdata/configs/envs/staging", true},
// 		{"testdata/configs/uninitialized", false},
// 	})
// 	require.NoError(t, err)
// 	assert.Equal(t, 5, len(got))
// 	assert.Contains(t, got, workspace{"default", Module{"testdata/configs/envs/dev", true}})
// 	assert.Contains(t, got, workspace{"non-default-1", Module{"testdata/configs/envs/dev", true}})
// 	assert.Contains(t, got, workspace{"non-default-2", Module{"testdata/configs/envs/dev", true}})
// 	assert.Contains(t, got, workspace{"default", Module{"testdata/configs/envs/prod", true}})
// 	assert.Contains(t, got, workspace{"default", Module{"testdata/configs/envs/staging", true}})
// }
//

func TestWorkspace_TerraformEnv(t *testing.T) {
	mod := module.New("a/b/c")
	ws := New(mod.Resource, "dev")

	assert.Equal(t, "TF_WORKSPACE=dev", ws.TerraformEnv())
}
