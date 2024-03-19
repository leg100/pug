package app

import (
	"testing"

	"github.com/peterbourgon/ff/v4"
	"github.com/stretchr/testify/assert"
)

func TestStart_help(t *testing.T) {
	// Short form
	err := Start([]string{"-h"})
	assert.ErrorIs(t, err, ff.ErrHelp)

	// Long form
	err = Start([]string{"--help"})
	assert.ErrorIs(t, err, ff.ErrHelp)
}
