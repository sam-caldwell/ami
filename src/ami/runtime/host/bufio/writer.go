package bufio

import (
    gobufio "bufio"
    "bytes"
    "io"
)

// Writer provides buffered writing to a destination. Flush must be called
// to make written data visible on the destination.
type Writer struct {
    w    *gobufio.Writer
    sink *bytes.Buffer // optional sink for tests
}

// NewWriter constructs a Writer from supported sinks: *bytes.Buffer, io.Writer.
func NewWriter(dst any) (Writer, error) {
    switch v := dst.(type) {
    case *bytes.Buffer:
        if v == nil { return Writer{}, ErrInvalidArg }
        return Writer{w: gobufio.NewWriter(v), sink: v}, nil
    case io.Writer:
        if v == nil { return Writer{}, ErrInvalidArg }
        return Writer{w: gobufio.NewWriter(v)}, nil
    default:
        return Writer{}, ErrInvalidArg
    }
}

// Write appends data from OwnedBytes to the buffer; caller may Release p after return.
func (wr Writer) Write(p OwnedBytes) (int, error) {
    if wr.w == nil { return 0, ErrInvalidHandle }
    b, err := p.Bytes()
    if err != nil { return 0, err }
    return wr.w.Write(b)
}

// Flush flushes the internal buffer to the destination.
func (wr Writer) Flush() error {
    if wr.w == nil { return ErrInvalidHandle }
    return wr.w.Flush()
}

