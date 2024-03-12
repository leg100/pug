package module

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindModules(t *testing.T) {
	got, err := findModules("../testdata/configs")
	require.NoError(t, err)

	assert.Equal(t, 5, len(got))
	assert.Contains(t, got, "envs/dev")
	assert.Contains(t, got, "envs/prod")
	assert.Contains(t, got, "envs/staging")
	assert.Contains(t, got, "uninitialized")
	assert.Contains(t, got, "plan_forever")
}
