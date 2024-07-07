package module

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTerragruntGraph(t *testing.T) {
	buf := bytes.NewBufferString(terragruntGraphDigraph)
	got, err := parseTerragruntGraph(buf)
	require.NoError(t, err)

	// Should find 5 modules
	assert.Len(t, got, 5)

	if assert.Contains(t, got, "root/backend-app") {
		if assert.Len(t, got["root/backend-app"], 3) {
			assert.Contains(t, got["root/backend-app"], "root/mysql")
			assert.Contains(t, got["root/backend-app"], "root/redis")
			assert.Contains(t, got["root/backend-app"], "root/vpc")
		}
	}
	if assert.Contains(t, got, "root/frontend-app") {
		if assert.Len(t, got["root/frontend-app"], 2) {
			assert.Contains(t, got["root/frontend-app"], "root/backend-app")
			assert.Contains(t, got["root/frontend-app"], "root/vpc")
		}
	}
	if assert.Contains(t, got, "root/mysql") {
		if assert.Len(t, got["root/mysql"], 1) {
			assert.Contains(t, got["root/mysql"], "root/vpc")
		}
	}
	if assert.Contains(t, got, "root/redis") {
		if assert.Len(t, got["root/redis"], 1) {
			assert.Contains(t, got["root/redis"], "root/vpc")
		}
	}
	if assert.Contains(t, got, "root/vpc") {
		assert.Empty(t, got["root/vpc"])
	}
}
