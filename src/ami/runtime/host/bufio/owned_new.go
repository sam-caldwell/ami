package bufio

// newOwnedBytes creates a new OwnedBytes handle from a copy of p.
func newOwnedBytes(p []byte) OwnedBytes {
    if len(p) == 0 {
        return OwnedBytes{b: nil, valid: true}
    }
    q := make([]byte, len(p))
    copy(q, p)
    return OwnedBytes{b: q, valid: true}
}

