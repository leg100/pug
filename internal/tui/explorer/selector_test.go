package explorer

import (
	"testing"

	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestSelector_isSelected(t *testing.T) {
	mod1 := moduleNode{id: resource.NewMonotonicID(resource.Module)}
	mod2 := moduleNode{id: resource.NewMonotonicID(resource.Module)}

	s := selector{selections: make(map[resource.ID]struct{})}
	s.add(mod1)

	assert.True(t, s.isSelected(mod1))
	assert.False(t, s.isSelected(mod2))
}

func TestSelector_reindex(t *testing.T) {
	mod1 := moduleNode{id: resource.NewMonotonicID(resource.Module)}
	mod2 := moduleNode{id: resource.NewMonotonicID(resource.Module)}

	s := selector{selections: make(map[resource.ID]struct{})}
	s.add(mod1)
	s.add(mod2)

	assert.True(t, s.isSelected(mod1))
	assert.True(t, s.isSelected(mod2))

	s.reindex([]node{mod1})

	assert.True(t, s.isSelected(mod1))
	assert.False(t, s.isSelected(mod2))
}
