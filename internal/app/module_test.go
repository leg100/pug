package app

import (
	"strings"
	"testing"
)

func TestModule(t *testing.T) {
	tm := setup(t)

	// Expect module to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")

	// Expect to see successful init message
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!")
	})
}
