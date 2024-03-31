package table

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testVar struct{}

var (
	ent1 = testVar{}
	ent2 = testVar{}
	ent3 = testVar{}
)

func setupTest() Model[int, testVar] {
	renderer := func(v testVar, s lipgloss.Style) RenderedRow { return nil }
	tbl := New[int, testVar](nil, renderer, 0, 0)
	tbl.SetItems(map[int]testVar{
		0: ent1,
		1: ent2,
		2: ent3,
	})
	return tbl
}

func TestTable_Highlighted(t *testing.T) {
	tbl := setupTest()

	got, ok := tbl.Highlighted()
	require.True(t, ok)

	assert.Equal(t, ent1, got.Value)
}

func TestTable_ToggleSelection(t *testing.T) {
	tbl := setupTest()

	tbl.ToggleSelection()

	assert.Len(t, tbl.Selected, 1)
	assert.Equal(t, tbl.Selected[0], ent1)
}
