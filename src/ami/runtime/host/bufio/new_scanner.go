package bufio

import (
    gobufio "bufio"
)

// NewScanner constructs a Scanner from a Reader.
func NewScanner(r Reader) (Scanner, error) {
    if r.rdr == nil { return Scanner{}, ErrInvalidArg }
    sc := gobufio.NewScanner(r.rdr)
    // Default split is ScanLines (deterministic); keep default bufsize
    return Scanner{s: sc}, nil
}

