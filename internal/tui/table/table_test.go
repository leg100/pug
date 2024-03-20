package table

import (
	"testing"

	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEntity struct {
	resource.Resource
}

const testEntityKind resource.Kind = 10

var (
	ent1 = &testEntity{Resource: resource.New(testEntityKind, "", nil)}
	ent2 = &testEntity{Resource: resource.New(testEntityKind, "", nil)}
	ent3 = &testEntity{Resource: resource.New(testEntityKind, "", nil)}
)

func setupTest() Model[*testEntity] {
	// setup table
	cellFunc := func(e *testEntity) []Cell { return nil }
	tbl := New[*testEntity](nil, cellFunc, 0, 0)
	tbl.SetItems([]*testEntity{ent1, ent2, ent3})
	return tbl
}

func TestTable_Highlighted(t *testing.T) {
	tbl := setupTest()

	got, ok := tbl.Highlighted()
	require.True(t, ok)

	assert.Equal(t, ent1, got)
}

func TestTable_ToggleSelection(t *testing.T) {
	tbl := setupTest()

	tbl.ToggleSelection()

	assert.Len(t, tbl.Selected, 1)
	assert.Contains(t, tbl.Selected, ent1.ID())
	assert.Equal(t, tbl.Selected[ent1.ID()], ent1)
}
