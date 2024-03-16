package run

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ParsePlanReport(t *testing.T) {
	logs, err := os.ReadFile("testdata/plan_with_changes.txt")
	require.NoError(t, err)

	changed, got, err := parsePlanReport(string(logs))
	require.NoError(t, err)

	assert.True(t, changed)
	want := report{
		Additions:    1,
		Changes:      0,
		Destructions: 1,
	}
	assert.Equal(t, want, got)
}

func Test_ParsePlanReport_OutputChanges(t *testing.T) {
	logs, err := os.ReadFile("./testdata/plan_output_changes.txt")
	require.NoError(t, err)

	changed, got, err := parsePlanReport(string(logs))
	require.NoError(t, err)

	// no resource changes, but the outputs did change, so should be true.
	assert.True(t, changed)
	want := report{
		Additions:    0,
		Changes:      0,
		Destructions: 0,
	}
	assert.Equal(t, want, got)
}

func Test_ParsePlanReport_NoChanges(t *testing.T) {
	logs, err := os.ReadFile("./testdata/plan_no_changes.txt")
	require.NoError(t, err)

	changed, got, err := parsePlanReport(string(logs))
	require.NoError(t, err)

	assert.False(t, changed)
	want := report{
		Additions:    0,
		Changes:      0,
		Destructions: 0,
	}

	assert.Equal(t, want, got)
}

func Test_ParseApplyReport(t *testing.T) {
	logs, err := os.ReadFile("testdata/apply.txt")
	require.NoError(t, err)

	got, err := parseApplyReport(string(logs))

	require.NoError(t, err)
	want := report{
		Additions:    1,
		Changes:      0,
		Destructions: 0,
	}
	assert.Equal(t, want, got)
}

func Test_ParseApplyReport_NoChanges(t *testing.T) {
	logs, err := os.ReadFile("testdata/apply_no_changes.txt")
	require.NoError(t, err)

	got, err := parseApplyReport(string(logs))
	require.NoError(t, err)

	want := report{
		Additions:    0,
		Changes:      0,
		Destructions: 0,
	}
	assert.Equal(t, want, got)
}
