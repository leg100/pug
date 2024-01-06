package testutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func ChTempDir(t *testing.T, path string) {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(path)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := os.Chdir(wd)
		require.NoError(t, err)
	})
}
