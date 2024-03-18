package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStart_help(t *testing.T) {
	err := Start([]string{"-h"})
	require.NoError(t, err)
}
