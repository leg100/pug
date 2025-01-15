package logging

import (
	"testing"

	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestReferenceUpdater(t *testing.T) {
	res := &fakeResource{MonotonicID: resource.NewMonotonicID(resource.Module)}
	updater := &ReferenceUpdater[*fakeResource]{
		Getter: &fakeResourceGetter{res: res},
		Name:   "fake",
		Field:  "FakeResourceID",
	}

	t.Run("replace resource id with resource", func(t *testing.T) {
		args := []any{"fake", res.MonotonicID}
		got := updater.UpdateArgs(args...)

		want := []any{"fake", res}
		assert.Equal(t, want, got)
	})

	t.Run("add resource when referenced from struct with pointer field", func(t *testing.T) {
		type logMsgArg struct {
			FakeResourceID *resource.MonotonicID
		}

		args := []any{"arg1", logMsgArg{FakeResourceID: &res.MonotonicID}}
		got := updater.UpdateArgs(args...)

		want := append(args, "fake", res)
		assert.Equal(t, want, got)
	})

	t.Run("add resource when referenced from struct with non-pointer field", func(t *testing.T) {
		type logMsgArg struct {
			FakeResourceID resource.MonotonicID
		}

		args := []any{"arg1", logMsgArg{FakeResourceID: res.MonotonicID}}
		got := updater.UpdateArgs(args...)

		want := append(args, "fake", res)
		assert.Equal(t, want, got)
	})

	t.Run("handle nil pointer from struct", func(t *testing.T) {
		type logMsgArg struct {
			FakeResourceID *resource.MonotonicID
		}

		args := []any{"arg1", logMsgArg{FakeResourceID: nil}}
		got := updater.UpdateArgs(args...)

		assert.Equal(t, got, got)
	})
}

type fakeResource struct {
	resource.MonotonicID
}

type fakeResourceGetter struct {
	res *fakeResource
}

func (f *fakeResourceGetter) Get(resource.ID) (*fakeResource, error) {
	return f.res, nil
}
