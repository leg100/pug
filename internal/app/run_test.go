package app

import (
	"testing"
)

func TestRun(t *testing.T) {
	tm := setup(t)

	// Initialize and apply run on modules/a
	initAndApplyModuleA(t, tm)
}
