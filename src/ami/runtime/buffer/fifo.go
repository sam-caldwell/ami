package buffer

import "sync"

// FIFOQueue implements a bounded FIFO queue with backpressure policies.
type FIFOQueue struct {
    MinCapacity int
    MaxCapacity int
    Backpressure string // block|dropOldest|dropNewest|shuntOldest|shuntNewest

    mu sync.Mutex
    q  []any
    pushN int
    popN  int
    dropN int
    fullN int
}

func NewFIFO(min, max int, bp string) *FIFOQueue {
    if min < 0 { min = 0 }
    if max < 0 { max = 0 }
    return &FIFOQueue{MinCapacity: min, MaxCapacity: max, Backpressure: bp, q: make([]any, 0)}
}

func (f *FIFOQueue) Push(v any) error {
    f.mu.Lock(); defer f.mu.Unlock()
    f.pushN++
    cap := f.MaxCapacity
    if cap > 0 && len(f.q) >= cap {
        switch f.Backpressure {
        case "block":
            f.fullN++
            return ErrFull
        case "dropOldest", "shuntOldest":
            if len(f.q) > 0 {
                f.q = f.q[1:]
                f.dropN++
            }
        case "dropNewest", "shuntNewest":
            f.dropN++
            return nil
        default:
            f.dropN++
            return nil
        }
    }
    f.q = append(f.q, v)
    return nil
}

func (f *FIFOQueue) Pop() (any, bool) {
    f.mu.Lock(); defer f.mu.Unlock()
    if len(f.q) == 0 { return nil, false }
    v := f.q[0]
    f.q = f.q[1:]
    f.popN++
    return v, true
}

func (f *FIFOQueue) Len() int { f.mu.Lock(); defer f.mu.Unlock(); return len(f.q) }
func (f *FIFOQueue) Counters() (push, pop, drop, full int) { f.mu.Lock(); defer f.mu.Unlock(); return f.pushN, f.popN, f.dropN, f.fullN }

