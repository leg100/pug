package task

import (
	"io"
	"sync"
)

// buffer captures task output and makes it available to multiple readers,
// providing the following functionality:
//
// * readers receive all of the output, from the beginning of the task through
// to its termination.
// * block reader if there is nothing more to read but the task hasn't yet
// finished.
type buffer struct {
	buf     []byte
	clients []*bufferClient
	mu      sync.Mutex

	// more signals when there is further task output and when the task has
	// finished.
	more chan struct{}
}

func newBuffer() *buffer {
	return &buffer{
		more: make(chan struct{}, 1),
	}
}

func (b *buffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buf = append(b.buf, p...)
	for _, client := range b.clients {
		client.mu.Lock()
		client.buf = append(client.buf, p...)
		client.mu.Unlock()

		// inform client there is more output
		select {
		case client.more <- struct{}{}:
		default:
		}
	}
	return len(p), nil
}

// NewReader returns a copy of the buffer to read from.
func (b *buffer) NewReader() io.Reader {
	client := &bufferClient{
		buf:  b.buf,
		more: b.more,
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.clients = append(b.clients, client)

	return client
}

func (b *buffer) Close() {
	close(b.more)
}

type bufferClient struct {
	buf    []byte
	offset int
	// mu guards access to buf
	mu   sync.Mutex
	more chan struct{}
}

func (c *bufferClient) Read(p []byte) (int, error) {
	if n := c.read(p); n > 0 {
		return n, nil
	}
	_, ok := <-c.more
	if n := c.read(p); n > 0 {
		return n, nil
	}
	if ok {
		return 0, nil
	} else {
		// TODO: handle case of len(p) == 0
		return 0, io.EOF
	}
}

func (c *bufferClient) read(p []byte) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	n := copy(p, c.buf[c.offset:])
	c.offset += n
	return n
}
