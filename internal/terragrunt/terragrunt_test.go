package terragrunt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindTerragruntModules(t *testing.T) {
	got, err := FindModules()
	require.NoError(t, err)

	// Should find 4 modules
	assert.Len(t, got, 4)

	// Should find backend-app module, which should have three dependencies
	if assert.Contains(t, got, "root/backend-app") {
		assert.Contains(t, got["root/backend-app"], "root/mysql")
		assert.Contains(t, got["root/backend-app"], "root/redis")
		assert.Contains(t, got["root/backend-app"], "root/vpc")
	}

	if assert.Contains(t, got, "root/frontend-app") {
		assert.Contains(t, got["root/frontend-app"], "root/backend-app")
		assert.Contains(t, got["root/frontend-app"], "root/vpc")
	}

	if assert.Contains(t, got, "root/mysql") {
		assert.Contains(t, got["root/mysql"], "root/vpc")
	}

	if assert.Contains(t, got, "root/redis") {
		assert.Contains(t, got["root/redis"], "root/vpc")
	}
}
