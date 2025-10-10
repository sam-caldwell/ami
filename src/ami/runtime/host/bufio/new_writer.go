package bufio

import (
    gobufio "bufio"
    "bytes"
    "io"
)

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

