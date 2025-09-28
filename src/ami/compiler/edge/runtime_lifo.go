package edge

import "sync"

// LIFOStack is a runtime buffer implementing LIFO (stack) semantics with optional bounds and backpressure.
type LIFOStack struct {
    spec LIFO
    mu   sync.Mutex
    buf  []any
    pushN int
    popN  int
    dropN int
    fullN int
}

func NewLIFO(spec LIFO) (*LIFOStack, error) {
    if err := spec.Validate(); err != nil { return nil, err }
    return &LIFOStack{spec: spec, buf: make([]any, 0)}, nil
}

// Push pushes v honoring backpressure when bounded and full.
func (s *LIFOStack) Push(v any) error {
    s.mu.Lock(); defer s.mu.Unlock()
    s.pushN++
    cap := s.spec.MaxCapacity
    if cap > 0 && len(s.buf) >= cap {
        switch s.spec.Backpressure {
        case "block":
            s.fullN++
            return ErrFull
        case "dropOldest", "shuntOldest":
            // drop bottom element (oldest) to make room
            copy(s.buf[0:], s.buf[1:])
            s.buf = s.buf[:len(s.buf)-1]
            s.dropN++
        case "dropNewest", "shuntNewest":
            // drop incoming element
            s.dropN++
            return nil
        default:
            s.dropN++
            return nil
        }
    }
    s.buf = append(s.buf, v)
    return nil
}

// Pop pops the newest element (top of stack). ok=false when empty.
func (s *LIFOStack) Pop() (v any, ok bool) {
    s.mu.Lock(); defer s.mu.Unlock()
    if len(s.buf) == 0 { return nil, false }
    i := len(s.buf) - 1
    v = s.buf[i]
    s.buf = s.buf[:i]
    s.popN++
    return v, true
}

func (s *LIFOStack) Len() int { s.mu.Lock(); defer s.mu.Unlock(); return len(s.buf) }

// Counters returns push, pop, drop, and full counts.
func (s *LIFOStack) Counters() (push, pop, drop, full int) {
    s.mu.Lock(); defer s.mu.Unlock()
    return s.pushN, s.popN, s.dropN, s.fullN
}
