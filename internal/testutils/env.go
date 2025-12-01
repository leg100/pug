package testutils

import (
	"os"
	"strings"
	"testing"
)

// ResetEnv unsets all environment variables for the duration of a test.
func ResetEnv(t *testing.T) {
	for _, env := range os.Environ() {
		k, v, _ := strings.Cut(env, "=")
		os.Unsetenv(k)
		t.Cleanup(func() {
			os.Setenv(k, v)
		})
	}
}
