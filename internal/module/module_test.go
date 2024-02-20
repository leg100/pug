package module

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindModules(t *testing.T) {
	got, err := FindModules("../testdata/configs")
	require.NoError(t, err)
	assert.Equal(t, 4, len(got))
	//assert.Equal(t, "", got[0].Path)
}

//func TestModule_init(t *testing.T) {
//	mod := Module{"testdata/configs/envs/dev", true}
//	err := mod.init(NewRunner(1))
//	require.NoError(t, err)
//}
