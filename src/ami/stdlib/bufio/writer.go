package bufio

import (
	stdbufio "bufio"
	stdio "io"
)

// Writer wraps a stdlib bufio.Writer with explicit buffer size and flush semantics.
type Writer struct{ w *stdbufio.Writer }

// NewWriterSize creates a new Writer with explicit buffer size. Size must be > 0.
func NewWriterSize(w stdio.Writer, size int) (*Writer, error) {
	if size <= 0 {
		return nil, ErrInvalidBufferSize
	}
	return &Writer{w: stdbufio.NewWriterSize(w, size)}, nil
}

// Write writes data into the internal buffer in chunks that do not exceed the
// configured buffer capacity. When the buffer becomes full and more input
// remains, Write flushes once before continuing. This guarantees observable
// partial flush behavior when the input exceeds the buffer size.
func (w *Writer) Write(p []byte) (int, error) {
	// Fast path: fits entirely without filling the buffer
	if len(p) <= w.w.Available() {
		return w.w.Write(p)
	}
	written := 0
	rem := p
	for len(rem) > 0 {
		avail := w.w.Available()
		if avail == 0 {
			if err := w.w.Flush(); err != nil {
				return written, err
			}
			avail = w.w.Available()
		}
		chunk := avail
		if len(rem) < chunk {
			chunk = len(rem)
		}
		n, err := w.w.Write(rem[:chunk])
		written += n
		rem = rem[n:]
		if err != nil {
			return written, err
		}
		// If we exactly filled the buffer and more data remains, flush now to
		// make the partial write visible to the underlying writer.
		if len(rem) > 0 && w.w.Available() == 0 {
			if err := w.w.Flush(); err != nil {
				return written, err
			}
		}
	}
	return written, nil
}

// Flush writes any buffered data to the underlying io.Writer.
func (w *Writer) Flush() error { return w.w.Flush() }
