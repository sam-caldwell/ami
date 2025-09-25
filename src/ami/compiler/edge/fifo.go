package edge

import (
    "errors"
    "sync"
    "sync/atomic"
)

// FIFO is a declarative, compiler-generated first-in, first-out edge.
// Codegen emits a high-performance MPMC ring buffer specialized to the payload
// type and capacity bounds.
type FIFO struct {
	MinCapacity  int
	MaxCapacity  int
	Backpressure BackpressurePolicy
	TypeName     string // payload type name for analysis/codegen
    // runtime buffer (scaffold for tests and harness)
    mu sync.Mutex
    q  []any
    n  uint32
}

func (f *FIFO) Kind() string { return "edge.FIFO" }

func (f *FIFO) Validate() error {
	if f.MinCapacity < 0 {
		return errors.New("minCapacity must be >= 0")
	}
	if f.MaxCapacity < f.MinCapacity {
		return errors.New("maxCapacity must be >= minCapacity")
	}
	switch f.Backpressure {
	case "", BackpressureBlock, BackpressureDrop:
	default:
		return errors.New("invalid backpressure policy")
	}
	return nil
}

// Count returns the number of queued events.
// A nil receiver is treated as empty and returns 0.
func (f *FIFO) Count() uint32 {
    if f == nil { return 0 }
    return atomic.LoadUint32(&f.n)
}

// Push appends an event to the FIFO. When capacity is bounded and the buffer
// is full, behavior depends on backpressure:
//  - block: returns ErrFull (would block in real runtime)
//  - drop: drops the oldest event to make room, then appends
func (f *FIFO) Push(e Event) error {
    f.mu.Lock()
    defer f.mu.Unlock()
    // bounded
    if f.MaxCapacity > 0 && len(f.q) >= f.MaxCapacity {
        switch f.Backpressure {
        case BackpressureBlock:
            return ErrFull
        case BackpressureDrop:
            // drop oldest (front); net count stays the same after append
            if len(f.q) > 0 {
                f.q = f.q[1:]
            }
        }
    } else {
        // increase queue size by 1
        atomic.AddUint32(&f.n, 1)
    }
    f.q = append(f.q, any(e))
    return nil
}

// Pop removes and returns the oldest event. ok=false when empty.
func (f *FIFO) Pop() (Event, bool) {
    f.mu.Lock()
    defer f.mu.Unlock()
    if len(f.q) == 0 { var z Event; return z, false }
    v := f.q[0]
    f.q = f.q[1:]
    atomic.AddUint32(&f.n, ^uint32(0)) // -1
    if ev, ok := v.(Event); ok { return ev, true }
    var z Event
    return z, false
}
