package gpu

// Pipeline represents a Metal compute pipeline.
type Pipeline struct{
    valid bool
    pipeId int
}

func (p *Pipeline) Release() error {
    if p == nil || !p.valid { return ErrInvalidHandle }
    if p.pipeId > 0 { metalReleasePipeline(p.pipeId) }
    p.valid = false
    p.pipeId = 0
    return nil
}

