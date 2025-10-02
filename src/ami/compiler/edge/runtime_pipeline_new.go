package edge

func NewPipelineBuffer() *PipelineBuffer {
    q, _ := NewFIFO(FIFO{MaxCapacity: 0}) // unbounded
    return &PipelineBuffer{fifo: q}
}

