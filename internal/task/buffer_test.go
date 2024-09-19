package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuffer_NewReader tests creating multiple readers and demonstrating that
// they each get their own copy of the underlying data.
func TestBuffer_NewReader(t *testing.T) {
	t.Parallel()

	buf := newBuffer()
	_, err := buf.Write([]byte("hello world"))
	require.NoError(t, err)

	r1 := buf.NewReader()
	r2 := buf.NewReader()

	got := make([]byte, len("hello world"))

	_, err = r1.Read(got)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(got))

	_, err = r2.Read(got)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(got))
}

func TestBuffer_Stream(t *testing.T) {
	t.Parallel()

	buf := newBuffer()
	ch := buf.Stream()

	_, err := buf.Write([]byte("hello"))
	require.NoError(t, err)

	got := <-ch
	assert.Equal(t, "hello", string(got))

	_, err = buf.Write([]byte("world"))
	require.NoError(t, err)

	got = <-ch
	assert.Equal(t, "world", string(got))

	buf.Close()

	got = <-ch
	assert.Nil(t, got)
}
