package testutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func ChTempDir(t *testing.T) {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)
	os.Chdir(t.TempDir())
	t.Cleanup(func() {
		os.Chdir(wd)
	})
}
