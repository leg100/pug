package resource

import (
	"fmt"
	"sync"

	"golang.org/x/exp/maps"
)

// Table is an in-memory database table that emits events upon changes.
type Table[T any] struct {
	rows map[ID]T
	mu   sync.RWMutex

	pub Publisher[T]
}

func NewTable[T any](pub Publisher[T]) *Table[T] {
	return &Table[T]{
		rows: make(map[ID]T),
		pub:  pub,
	}
}

func (t *Table[T]) Add(id ID, row T) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.rows[id] = row
	t.pub.Publish(CreatedEvent, row)
}

func (t *Table[T]) Update(id ID, updater func(existing T) error) (T, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Retrieve row and update
	row, ok := t.rows[id]
	if !ok {
		return *new(T), ErrNotFound
	}
	if err := updater(row); err != nil {
		return *new(T), err
	}
	t.rows[id] = row

	t.pub.Publish(UpdatedEvent, row)
	return row, nil
}

func (t *Table[T]) Delete(id ID) {
	t.mu.Lock()
	defer t.mu.Unlock()

	row := t.rows[id]
	delete(t.rows, id)
	t.pub.Publish(DeletedEvent, row)
}

func (t *Table[T]) Get(id ID) (T, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	row, ok := t.rows[id]
	if !ok {
		return *new(T), fmt.Errorf("%s: %w", id, ErrNotFound)
	}
	return row, nil
}

func (t *Table[T]) List() []T {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return maps.Values(t.rows)
}
