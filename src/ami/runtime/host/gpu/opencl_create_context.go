package gpu

// OpenCLCreateContext creates an OpenCL context (stub: unavailable).
func OpenCLCreateContext(p Platform) (Context, error) {
    if p.Vendor == "" && p.Name == "" && p.Version == "" {
        return Context{}, ErrInvalidHandle
    }
    return Context{}, ErrUnavailable
}

