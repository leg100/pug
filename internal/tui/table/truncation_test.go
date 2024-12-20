package table

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateLeft_NoTruncation(t *testing.T) {
	path := "/a/b/c/d/e/f"
	got := TruncateLeft(path, 99, "…")
	assert.Equal(t, path, got)
}

func TestTruncateLeft_Truncate(t *testing.T) {
	path := "/a/b/c/d/e/f"
	got := TruncateLeft(path, 5, "…")
	assert.Equal(t, "…/e/f", got)
}
