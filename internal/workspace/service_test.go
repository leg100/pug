package workspace

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_parseList(t *testing.T) {
	r := strings.NewReader(`
  default
* foo
  bar
`)

	got, current, err := parseList(r)
	require.NoError(t, err)

	assert.Len(t, got, 3)
	assert.Contains(t, got, "default")
	assert.Contains(t, got, "foo")
	assert.Contains(t, got, "bar")

	assert.Equal(t, "foo", current)
}
