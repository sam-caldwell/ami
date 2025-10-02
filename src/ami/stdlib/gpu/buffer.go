package gpu

// Buffer represents device memory.
type Buffer struct {
    backend string
    n       int
    valid   bool
    bufId   int
}

// Release releases the buffer. Returns ErrInvalidHandle for zero or released.
func (b *Buffer) Release() error {
    if b == nil || !b.valid {
        return ErrInvalidHandle
    }
    if b.backend == "metal" && b.bufId > 0 {
        metalFreeBufferByID(b.bufId)
    }
    b.backend = ""
    b.n = 0
    b.bufId = 0
    b.valid = false
    return nil
}

