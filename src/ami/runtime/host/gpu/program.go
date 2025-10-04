package gpu

// Program represents an OpenCL program.
type Program struct{
    valid bool
}

func (p *Program) Release() error {
    if p == nil || !p.valid { return ErrInvalidHandle }
    p.valid = false
    return nil
}

