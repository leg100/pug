package workspace

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOverallCost(t *testing.T) {
	in, err := os.Open("./testdata/infracost.table")
	require.NoError(t, err)

	got, err := parseInfracostOutput(in)
	require.NoError(t, err)

	assert.Equal(t, "$9.27", got)
}
