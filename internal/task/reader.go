package task

import "io"

// reader wraps the task buffer, blocking reads when there is nothing currently
// to be read and only returning an io.EOF once the the task buffer has been
// read in its entirety and the task has finished.
type reader struct {
	buf    *buffer
	offset int
}

// Read blocks when there is nothing to be read.
func (b *reader) Read(p []byte) (int, error) {
	if n, err := b.readWithLock(p); n > 0 {
		return n, err
	}
	// buffer is empty: wait til something is written
	<-b.buf.avail

	// now read again
	return b.readWithLock(p)
}

func (b *reader) readWithLock(p []byte) (int, error) {
	b.buf.mu.Lock()
	defer b.buf.mu.Unlock()

	byts := b.buf.buf.Bytes()
	n := copy(p, byts[b.offset:])
	b.offset += n
	// return io.EOF if everything to be read has been read.
	if b.offset == b.buf.buf.Len() && b.buf.closed {
		return n, io.EOF
	}
	return n, nil
}
