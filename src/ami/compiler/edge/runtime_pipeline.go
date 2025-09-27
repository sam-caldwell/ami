package edge

// PipelineBuffer is a simple unbounded FIFO used to connect pipelines in runtime scaffolding.
type PipelineBuffer struct { fifo *FIFOQueue }

func NewPipelineBuffer() *PipelineBuffer {
    q, _ := NewFIFO(FIFO{MaxCapacity: 0}) // unbounded
    return &PipelineBuffer{fifo: q}
}

func (p *PipelineBuffer) Push(v any) error { return p.fifo.Push(v) }
func (p *PipelineBuffer) Pop() (any, bool) { return p.fifo.Pop() }
func (p *PipelineBuffer) Len() int         { return p.fifo.Len() }

