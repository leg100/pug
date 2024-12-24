package app

import (
	"strings"
	"testing"
)

func TestExplorer_Filter(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Focus filter widget
	tm.Type("/")

	// Expect filter prompt
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Filter:")
	})

	// Filter to only show modules/a
	tm.Type("a")

	// Expect table title to show 1 module filtered out of a total 3.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "󰠱 1  0")
	})
}
