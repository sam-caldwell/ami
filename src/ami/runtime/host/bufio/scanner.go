package bufio

import (
    gobufio "bufio"
)

// Scanner wraps bufio.Scanner to provide simple line scanning with
// Owned-like Bytes() handle and deterministic limits.
type Scanner struct {
    s   *gobufio.Scanner
    tok []byte
}

// NewScanner constructs a Scanner from a Reader.
func NewScanner(r Reader) (Scanner, error) {
    if r.rdr == nil { return Scanner{}, ErrInvalidArg }
    sc := gobufio.NewScanner(r.rdr)
    // Default split is ScanLines (deterministic); keep default bufsize
    return Scanner{s: sc}, nil
}

// Scan advances the scanner to the next token. It copies token bytes into
// an internal buffer to ensure Bytes() returns an Owned copy independent of
// the underlying bufio.Scanner buffer.
func (sc *Scanner) Scan() bool {
    if sc == nil || sc.s == nil { return false }
    if !sc.s.Scan() { sc.tok = nil; return false }
    b := sc.s.Bytes()
    if len(b) == 0 { sc.tok = nil; return true }
    sc.tok = append(sc.tok[:0], b...)
    return true
}

// Text returns the current token as string.
func (sc *Scanner) Text() string {
    if sc == nil { return "" }
    if sc.tok == nil { return sc.s.Text() }
    return string(sc.tok)
}

// Bytes returns the current token as an OwnedBytes handle.
func (sc *Scanner) Bytes() OwnedBytes {
    if sc == nil { return newOwnedBytes(nil) }
    return newOwnedBytes(sc.tok)
}

// Err reports the first non-EOF error that was encountered by the Scanner.
func (sc *Scanner) Err() error {
    if sc == nil || sc.s == nil { return ErrInvalidHandle }
    return sc.s.Err()
}

