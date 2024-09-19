package task

import (
	"bytes"
	"io"
	"sync"
)

type buffer struct {
	buf *bytes.Buffer

	avail chan struct{}
	mu    sync.Mutex
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
	// Let streamers know there are now available bytes to be read.
	select {
	case b.avail <- struct{}{}:
	default:
	}
	return n, nil
}

// NewReader returns a copy of the buffer to read from.
func (b *buffer) NewReader() io.Reader {
	b.mu.Lock()
	defer b.mu.Unlock()

	r := new(bytes.Buffer)
	r.Write(b.buf.Bytes())
	return r
}

// Stream buffer as it is written to. The return channel is closed when the
// buffer is closed.
func (b *buffer) Stream() <-chan []byte {
	var (
		offset int
		ch     = make(chan []byte)
	)

	copyBytes := func() []byte {
		b.mu.Lock()
		byts := b.buf.Bytes()
		size := b.buf.Len() - offset
		dst := make([]byte, size)
		offset += copy(dst, byts[offset:])
		b.mu.Unlock()
		return dst
	}

	go func() {
		if b := copyBytes(); len(b) > 0 {
			ch <- b
		}
		for {
			_, ok := <-b.avail
			if b := copyBytes(); len(b) > 0 {
				ch <- b
			}
			if !ok {
				close(ch)
				return
			}
		}
	}()
	return ch
}

func (b *buffer) Close() {
	close(b.avail)
}
