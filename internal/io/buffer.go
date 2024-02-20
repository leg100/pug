package io

import (
	"bytes"
	"io"
	"sync"
)

// Buffer blocks the reader until there is something to read.
type Buffer struct {
	buf   *bytes.Buffer
	avail chan struct{}
	mu    sync.Mutex
}

func NewBuffer() *Buffer {
	return &Buffer{
		buf:   new(bytes.Buffer),
		avail: make(chan struct{}, 1),
	}
}

func (b *Buffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	n, err := b.buf.Write(p)
	if err != nil {
		return n, err
	}
	select {
	case b.avail <- struct{}{}:
	default:
	}
	return n, nil
}

// Read buffer contents into p. If the buffer is empty then Read blocks until it
// is non-empty or the buffer is closed.
func (b *Buffer) Read(p []byte) (int, error) {
	n, err := b.readWithLock(p)
	if err != nil {
		if err != io.EOF {
			// return non-EOF error
			return n, err
		}
	}
	if n > 0 {
		// something was read, so return it along with any EOF error.
		return n, err
	}
	// buffer is empty: wait til something is written
	<-b.avail

	// now read again
	return b.readWithLock(p)
}

func (b *Buffer) readWithLock(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	n, err := b.buf.Read(p)
	if err != nil {
		return n, err
	}
	if b.buf.Len() == 0 {
		// send io.EOF when buffer is empty (bytes.Buffer.Read does not return
		// an io.EOF until the second call).
		return n, io.EOF
	}
	return n, nil
}

func (b *Buffer) Close() {
	close(b.avail)
}
