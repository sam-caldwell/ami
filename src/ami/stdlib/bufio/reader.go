package bufio

import (
    stdbufio "bufio"
    stdio "io"
)

// Reader wraps a stdlib bufio.Reader with explicit buffer size.
type Reader struct { r *stdbufio.Reader }

// NewReaderSize creates a new Reader with explicit buffer size. Size must be > 0.
func NewReaderSize(rd stdio.Reader, size int) (*Reader, error) {
    if size <= 0 { return nil, ErrInvalidBufferSize }
    return &Reader{ r: stdbufio.NewReaderSize(rd, size) }, nil
}

// Read reads data into p.
func (r *Reader) Read(p []byte) (int, error) { return r.r.Read(p) }

