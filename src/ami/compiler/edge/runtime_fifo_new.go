package edge

func NewFIFO(spec FIFO) (*FIFOQueue, error) {
    if err := spec.Validate(); err != nil { return nil, err }
    return &FIFOQueue{spec: spec, buf: make([]any, 0)}, nil
}

