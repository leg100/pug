package testutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func ParseTime(t *testing.T, layout, s string) time.Time {
	got, err := time.Parse(layout, s)
	require.NoError(t, err)
	return got
}
