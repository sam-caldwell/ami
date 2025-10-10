package bufio

// OwnedBytes simulates an AMI Owned<slice<uint8>> handle in host tests.
// It owns a byte slice until Release() is called, after which it becomes invalid
// and double-release returns ErrAlreadyReleased.
type OwnedBytes struct {
    b     []byte
    valid bool
}

// newOwnedBytes moved to owned_new.go to satisfy single-declaration lint.

// Bytes returns the underlying slice. Caller must not modify after Release.
func (o *OwnedBytes) Bytes() ([]byte, error) {
    if o == nil || !o.valid {
        return nil, ErrInvalidHandle
    }
    return o.b, nil
}

// Len reports the length of the underlying slice.
func (o *OwnedBytes) Len() (int, error) {
    if o == nil || !o.valid { return 0, ErrInvalidHandle }
    return len(o.b), nil
}

// Release zeroizes the slice and invalidates the handle.
func (o *OwnedBytes) Release() error {
    if o == nil || !o.valid { return ErrAlreadyReleased }
    for i := range o.b { o.b[i] = 0 }
    o.b = nil
    o.valid = false
    return nil
}
