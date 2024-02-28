package task

import "io"

// reader provides a Read() that blocks when there is nothing to be read.
type reader struct {
	buf    *buffer
	offset int
}

// Read blocks when there is nothing to be read.
func (b *reader) Read(p []byte) (int, error) {
	// if error is non-nil it'll be io.EOF
	n, err := b.readWithLock(p)
	if n > 0 {
		// something was read, so return it along with err
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
	if b.offset == b.buf.buf.Len() {
		// reader has caught up with the end of the buffer, so send an io.EOF
		// now rather than blocking the caller on the next call.
		return n, io.EOF
	}
	return n, nil
}
