package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindWorkspaces(t *testing.T) {
	got, err := findWorkspaces([]module{
		{"testdata/configs/envs/dev", true},
		{"testdata/configs/envs/prod", true},
		{"testdata/configs/envs/staging", true},
		{"testdata/configs/uninitialized", false},
	})
	require.NoError(t, err)
	assert.Equal(t, 5, len(got))
	assert.Contains(t, got, workspace{"default", module{"testdata/configs/envs/dev", true}})
	assert.Contains(t, got, workspace{"non-default-1", module{"testdata/configs/envs/dev", true}})
	assert.Contains(t, got, workspace{"non-default-2", module{"testdata/configs/envs/dev", true}})
	assert.Contains(t, got, workspace{"default", module{"testdata/configs/envs/prod", true}})
	assert.Contains(t, got, workspace{"default", module{"testdata/configs/envs/staging", true}})
}
