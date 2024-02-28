package module

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindModules(t *testing.T) {
	got, err := findModules("../testdata/configs")
	require.NoError(t, err)
	assert.Equal(t, 4, len(got))
	//assert.Equal(t, "", got[0].Path)
}
