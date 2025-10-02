package buffer

func NewFIFO(min, max int, bp string) *FIFOQueue {
    if min < 0 { min = 0 }
    if max < 0 { max = 0 }
    return &FIFOQueue{MinCapacity: min, MaxCapacity: max, Backpressure: bp, q: make([]any, 0)}
}

