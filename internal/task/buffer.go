package task

import (
	"bytes"
	"io"
	"sync"
)

type buffer struct {
	buf     []byte
	clients []*io.PipeWriter
	mu      sync.Mutex
}

func newBuffer() *buffer {
	return &buffer{}
}

func (b *buffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buf = append(b.buf, p...)

	for _, client := range b.clients {
		_, _ = io.Copy(client, bytes.NewReader(p))
	}
	return len(p), nil
}

// NewReader returns a copy of the buffer to read from.
func (b *buffer) NewReader() io.Reader {
	r, w := io.Pipe()

	b.mu.Lock()
	b.clients = append(b.clients, w)
	b.mu.Unlock()

	go func() {
		_, _ = io.Copy(w, bytes.NewReader(b.buf))
	}()

	return r
}

func (b *buffer) Close() {
	for _, client := range b.clients {
		_ = client.Close()
	}
}
