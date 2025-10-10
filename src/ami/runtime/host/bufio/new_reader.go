package bufio

import (
    gobufio "bufio"
    "bytes"
    "io"
)

// NewReader constructs a Reader from supported sources: string, []byte, *bytes.Buffer, io.Reader.
func NewReader(src any) (Reader, error) {
    switch v := src.(type) {
    case string:
        b := bytes.NewBufferString(v)
        return Reader{rdr: gobufio.NewReader(b), buf: b}, nil
    case []byte:
        b := bytes.NewBuffer(append([]byte(nil), v...))
        return Reader{rdr: gobufio.NewReader(b), buf: b}, nil
    case *bytes.Buffer:
        if v == nil { return Reader{}, ErrInvalidArg }
        return Reader{rdr: gobufio.NewReader(v), buf: v}, nil
    case io.Reader:
        if v == nil { return Reader{}, ErrInvalidArg }
        return Reader{rdr: gobufio.NewReader(v)}, nil
    default:
        return Reader{}, ErrInvalidArg
    }
}

