package task

import (
	"io"
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
	buf.Close()

	r1 := buf.NewReader()
	r2 := buf.NewReader()

	got, err := io.ReadAll(r1)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(got))

	got, err = io.ReadAll(r2)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(got))
}

func TestBuffer_Stream(t *testing.T) {
	t.Parallel()

	buf := newBuffer()
	r1 := buf.NewReader()

	_, err := buf.Write([]byte("hello"))
	require.NoError(t, err)

	p := make([]byte, 5)
	_, err = r1.Read(p)
	require.NoError(t, err)

	assert.Equal(t, "hello", string(p))

	_, err = buf.Write([]byte(" world"))
	require.NoError(t, err)

	p = make([]byte, 6)
	_, err = r1.Read(p)
	require.NoError(t, err)

	assert.Equal(t, " world", string(p))

	buf.Close()

	got, err := io.ReadAll(r1)
	assert.Equal(t, 0, len(got))
	assert.NoError(t, err)
}
