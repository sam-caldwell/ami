package gpu

// Context represents a GPU execution context.
type Context struct {
    backend string
    valid   bool
    ctxId   int
}

// Release releases the context. Returns ErrInvalidHandle for zero or released.
func (c *Context) Release() error {
    if c == nil || !c.valid {
        return ErrInvalidHandle
    }
    // backend-specific teardown
    if c.backend == "metal" && c.ctxId > 0 {
        metalDestroyContextByID(c.ctxId)
    }
    c.backend = ""
    c.ctxId = 0
    c.valid = false
    return nil
}

