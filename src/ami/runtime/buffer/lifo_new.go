package buffer

func NewLIFO(min, max int, bp string) *LIFOStack {
    if min < 0 { min = 0 }
    if max < 0 { max = 0 }
    return &LIFOStack{MinCapacity: min, MaxCapacity: max, Backpressure: bp, s: make([]any, 0)}
}

