// Package iox provides small concurrency-safe I/O helpers used to avoid
// interleaved writes when multiple goroutines share a single underlying
// writer (e.g. os.Stdout).
package iox

import (
	"io"
	"sync"
)

// LockedLineWriter serializes writes to an underlying io.Writer so that
// concurrent producers never interleave partial lines. Every call to
// WriteLine writes its argument plus a trailing newline as one atomic,
// mutex-protected operation.
type LockedLineWriter struct {
	mu sync.Mutex
	w  io.Writer
}

// NewLockedLineWriter wraps w for line-atomic concurrent writes.
func NewLockedLineWriter(w io.Writer) *LockedLineWriter {
	return &LockedLineWriter{w: w}
}

// WriteLine writes line followed by a single '\n', holding the lock for the
// duration of both writes so no other goroutine's WriteLine call can
// interleave its bytes in between. line should NOT already contain a
// trailing newline.
func (lw *LockedLineWriter) WriteLine(line []byte) error {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	if _, err := lw.w.Write(line); err != nil {
		return err
	}
	_, err := lw.w.Write([]byte{'\n'})
	return err
}
