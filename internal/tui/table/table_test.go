package table

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

// setupTest sets up a table test with several rows. Each row is keyed with an
// int, and the row item is an int corresponding to the key, for ease of
// testing. The rows are sorted from lowest int to highest int.
func setupTest() Model[int, int] {
	renderer := func(v int) RenderedRow { return nil }
	tbl := New[int, int](nil, renderer, 0, 0).
		WithSortFunc(func(i, j int) int {
			if i < j {
				return -1
			}
			return 1
		})
	tbl.SetItems(map[int]int{
		0: 0,
		1: 1,
		2: 2,
		3: 3,
		4: 4,
		5: 5,
	})
	return tbl
}

func TestTable_Highlighted(t *testing.T) {
	tbl := setupTest()

	got, ok := tbl.Highlighted()
	require.True(t, ok)

	assert.Equal(t, 0, got.Value)
}

func TestTable_ToggleSelection(t *testing.T) {
	tbl := setupTest()

	tbl.ToggleSelection()

	assert.Len(t, tbl.Selected, 1)
	assert.Equal(t, 0, tbl.Selected[1])
}

func TestTable_SelectRange(t *testing.T) {
	tests := []struct {
		name     string
		selected []int
		cursor   int
		want     []int
	}{
		{
			name:     "select no range when nothing is selected, and cursor is on first row",
			selected: []int{},
			want:     []int{},
		},
		{
			name:     "select no range when nothing is selected, and cursor is on last row",
			selected: []int{},
			want:     []int{},
		},
		{
			name:     "select no range when cursor is on the only selected row",
			selected: []int{0},
			want:     []int{0},
		},
		{
			name:     "select all rows between selected top row and cursor on last row",
			selected: []int{0}, // first row
			cursor:   5,        // last row
			want:     []int{0, 1, 2, 3, 4, 5},
		},
		{
			name:     "select rows between selected top row and cursor in third row",
			selected: []int{0}, // first row
			cursor:   2,        // third row
			want:     []int{0, 1, 2},
		},
		{
			name:     "select rows between selected top row and cursor in third row, ignoring selected last row",
			selected: []int{0, 5}, // first and last row
			cursor:   2,           // third row
			want:     []int{0, 1, 2, 5},
		},
		{
			name:     "select rows between cursor in third row and selected last row",
			selected: []int{5}, // last row
			cursor:   2,        // third row
			want:     []int{2, 3, 4, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := setupTest()
			for _, key := range tt.selected {
				tbl.ToggleSelectionByKey(key)
			}
			tbl.cursor = tt.cursor

			tbl.SelectRange()

			got := maps.Keys(tbl.Selected)
			slices.Sort(got)
			assert.Equal(t, tt.want, got)
		})
	}
}
