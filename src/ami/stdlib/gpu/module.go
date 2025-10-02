package gpu

// Module represents a CUDA module (PTX/Cubin).
type Module struct{
    valid bool
}

func (m *Module) Release() error {
    if m == nil || !m.valid { return ErrInvalidHandle }
    m.valid = false
    return nil
}

