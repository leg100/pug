package task

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuffer_singleRead(t *testing.T) {
	buf := newBuffer()
	r := &reader{buf: buf}
	_, err := buf.Write([]byte("hello world"))
	require.NoError(t, err)

	got := make([]byte, len("hello world"))
	_, err = r.Read(got)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(got))
}

func TestBuffer_multiRead(t *testing.T) {
	buf := newBuffer()
	r := &reader{buf: buf}
	_, err := buf.Write([]byte("hello world"))
	require.NoError(t, err)

	got := make([]byte, 5)
	n, err := r.Read(got)
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "hello", string(got))

	got = make([]byte, 5)
	n, err = r.Read(got)
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, " worl", string(got))

	// mimic task completion
	buf.Close()

	got = make([]byte, 5)
	n, err = r.Read(got)
	require.Equal(t, io.EOF, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, "d\x00\x00\x00\x00", string(got))
}

func TestBuffer_multiWrite(t *testing.T) {
	buf := newBuffer()
	r := &reader{buf: buf}
	_, err := buf.Write([]byte("hello"))
	require.NoError(t, err)

	_, err = buf.Write([]byte(" world"))
	require.NoError(t, err)

	// mimic task completion
	buf.Close()

	got := make([]byte, 13)
	n, err := r.Read(got)
	require.Equal(t, err, io.EOF)
	assert.Equal(t, 11, n)
	assert.Equal(t, "hello world\x00\x00", string(got))
}

func TestBuffer_blockRead(t *testing.T) {
	buf := newBuffer()
	r := &reader{buf: buf}

	wait := make(chan struct{})
	go func() {
		// this'll block
		got := make([]byte, 5)
		n, err := r.Read(got)

		// now unblocked
		require.NoError(t, err)
		assert.Equal(t, 5, n)
		assert.Equal(t, "hello", string(got))

		close(wait)
	}()

	// wait a bit to give Read above a chance to get blocked
	time.Sleep(time.Second)

	// this'll unblock Read
	_, err := buf.Write([]byte("hello world"))
	require.NoError(t, err)

	// wait til blocked Read above has finished
	<-wait

	got := make([]byte, 5)
	n, err := r.Read(got)
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, " worl", string(got))

	// mimic task completion
	buf.Close()

	got = make([]byte, 5)
	n, err = r.Read(got)
	require.Equal(t, io.EOF, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, "d\x00\x00\x00\x00", string(got))
}
