package tui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgo(t *testing.T) {
	now := time.Now()

	assert.Equal(t, "50s ago", Ago(now, now.Add(-47*time.Second)))
	assert.Equal(t, "47h ago", Ago(now, now.Add(-47*time.Hour)))
}
