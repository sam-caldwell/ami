package edge

func NewLIFO(spec LIFO) (*LIFOStack, error) {
    if err := spec.Validate(); err != nil { return nil, err }
    return &LIFOStack{spec: spec, buf: make([]any, 0)}, nil
}

