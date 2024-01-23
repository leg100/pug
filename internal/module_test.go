package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindModules(t *testing.T) {
	got, err := findModules("./testdata")
	require.NoError(t, err)
	want := []module{
		{"testdata/configs/envs/dev", true},
		{"testdata/configs/envs/prod", true},
		{"testdata/configs/envs/staging", true},
		{"testdata/configs/uninitialized", false},
	}
	assert.Equal(t, want, got)
}

func TestModule_init(t *testing.T) {
	mod := module{"testdata/configs/envs/dev", true}
	err := mod.init(NewRunner(1))
	require.NoError(t, err)
}
