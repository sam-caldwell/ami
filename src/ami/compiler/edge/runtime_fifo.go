package edge

import "sync"

// FIFOQueue is a runtime buffer implementing FIFO semantics with optional bounds and backpressure.
type FIFOQueue struct {
    spec FIFO
    mu   sync.Mutex
    buf  []any
    // counters (protected by mu)
    pushN int
    popN  int
    dropN int
    fullN int
}

func NewFIFO(spec FIFO) (*FIFOQueue, error) {
    if err := spec.Validate(); err != nil { return nil, err }
    return &FIFOQueue{spec: spec, buf: make([]any, 0)}, nil
}

// Push enqueues v honoring backpressure policy when bounded and full.
func (q *FIFOQueue) Push(v any) error {
    q.mu.Lock(); defer q.mu.Unlock()
    q.pushN++
    cap := q.spec.MaxCapacity
    if cap > 0 && len(q.buf) >= cap {
        switch q.spec.Backpressure {
        case "block":
            q.fullN++
            return ErrFull
        case "dropOldest", "shuntOldest":
            // drop front element to make room
            copy(q.buf[0:], q.buf[1:])
            q.buf = q.buf[:len(q.buf)-1]
            q.dropN++
        case "dropNewest", "shuntNewest":
            // drop incoming element silently
            q.dropN++
            return nil
        default:
            // default to best-effort dropNewest
            q.dropN++
            return nil
        }
    }
    q.buf = append(q.buf, v)
    return nil
}

// Pop dequeues the oldest element. ok=false when empty.
func (q *FIFOQueue) Pop() (v any, ok bool) {
    q.mu.Lock(); defer q.mu.Unlock()
    if len(q.buf) == 0 { return nil, false }
    v = q.buf[0]
    copy(q.buf[0:], q.buf[1:])
    q.buf = q.buf[:len(q.buf)-1]
    q.popN++
    return v, true
}

// Len returns the number of buffered elements (for tests/metrics).
func (q *FIFOQueue) Len() int { q.mu.Lock(); defer q.mu.Unlock(); return len(q.buf) }

// Counters returns push, pop, drop, and full counts.
func (q *FIFOQueue) Counters() (push, pop, drop, full int) {
    q.mu.Lock(); defer q.mu.Unlock()
    return q.pushN, q.popN, q.dropN, q.fullN
}
