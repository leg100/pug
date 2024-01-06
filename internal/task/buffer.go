package task

import (
	"bytes"
	"sync"
)

// buffer lets downward implementations know when there is a write via a
// channel, and allows them, via a mutex, to copy bytes from the buffer.
type buffer struct {
	buf    *bytes.Buffer
	avail  chan struct{}
	closed bool
	mu     sync.Mutex
}

func newBuffer() *buffer {
	return &buffer{
		buf:   new(bytes.Buffer),
		avail: make(chan struct{}, 1),
	}
}

func (b *buffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	n, err := b.buf.Write(p)
	if err != nil {
		return n, err
	}
	// populate channel, skipping if already populated
	select {
	case b.avail <- struct{}{}:
	default:
	}
	return n, nil
}

func (b *buffer) Close() {
	close(b.avail)
	b.closed = true
}
