package state

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getResourcesFromFile(t *testing.T) {
	b, err := os.ReadFile("./testdata/state.json")
	require.NoError(t, err)

	var file StateFile
	err = json.Unmarshal(b, &file)
	require.NoError(t, err)

	got := getResourcesFromFile(file)

	assert.Len(t, got, 8)

	assert.Contains(t, got, ResourceAddress("random_pet.pet"))
	assert.Contains(t, got, ResourceAddress("random_integer.suffix"))
	assert.Contains(t, got, ResourceAddress("module.child1.random_pet.pet"))
	assert.Contains(t, got, ResourceAddress("module.child1.random_integer.suffix"))
	assert.Contains(t, got, ResourceAddress("module.child2.random_integer.suffix"))
	assert.Contains(t, got, ResourceAddress("module.child2.module.child3.random_integer.suffix"))
	assert.Contains(t, got, ResourceAddress("module.child2.module.child3.random_pet.pet"))

	assert.Equal(t, Tainted, got["module.child2.random_pet.pet"].Status)
}

func Test_getResourcesFromFile_empty(t *testing.T) {
	b, err := os.ReadFile("./testdata/state_empty.json")
	require.NoError(t, err)

	var file StateFile
	err = json.Unmarshal(b, &file)
	require.NoError(t, err)

	got := getResourcesFromFile(file)
	assert.Len(t, got, 0)
}
