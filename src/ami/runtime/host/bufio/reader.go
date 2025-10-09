package bufio

import (
    "bytes"
    gobufio "bufio"
    "io"
)

// Reader provides buffered reading over a fixed source, returning OwnedBytes
// handles for AMI-like semantics in tests.
type Reader struct {
    rdr  *gobufio.Reader
    buf  *bytes.Buffer // optional backing store when src is []byte or string
}

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

// Peek returns up to n bytes without advancing the reader. Returns fewer
// than n bytes if fewer are available.
func (r Reader) Peek(n int) (OwnedBytes, error) {
    if r.rdr == nil || n < 0 { return OwnedBytes{}, ErrInvalidArg }
    if n == 0 { return newOwnedBytes(nil), nil }
    p, _ := r.rdr.Peek(n) // go bufio.Peek returns fewer or error; we tolerate short reads
    // Copy to OwnedBytes to enforce ownership/immutability in tests
    return newOwnedBytes(p), nil
}

// Read returns up to n bytes and advances the reader.
func (r Reader) Read(n int) (OwnedBytes, error) {
    if r.rdr == nil || n < 0 { return OwnedBytes{}, ErrInvalidArg }
    if n == 0 { return newOwnedBytes(nil), nil }
    // Use Read to honor buffered behavior and advance
    buf := make([]byte, n)
    m, err := io.ReadAtLeast(r.rdr, buf, 1)
    if err == gobufio.ErrBufferFull || err == nil {
        // m bytes available; truncate to m
        return newOwnedBytes(buf[:m]), nil
    }
    if err == io.EOF || err == io.ErrUnexpectedEOF {
        if m > 0 { return newOwnedBytes(buf[:m]), nil }
        return newOwnedBytes(nil), nil
    }
    return OwnedBytes{}, err
}

// UnreadByte steps back one byte in the reader if possible.
func (r Reader) UnreadByte() error {
    if r.rdr == nil { return ErrInvalidHandle }
    return r.rdr.UnreadByte()
}

