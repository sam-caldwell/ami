package edge

import (
	"errors"
	"sync"
	"sync/atomic"
)

// LIFO is a declarative, compiler-generated last-in, first-out edge.
// Codegen emits a high-performance stack-like ring optimized per payload type.
type LIFO struct {
	MinCapacity  int
	MaxCapacity  int
	Backpressure BackpressurePolicy
	TypeName     string // payload type name for analysis/codegen
	// runtime buffer (stack) scaffold
	mu sync.Mutex
	q  []any
	n  uint32
}

func (l *LIFO) Kind() string { return "edge.LIFO" }

func (l *LIFO) Validate() error {
	if l.MinCapacity < 0 {
		return errors.New("minCapacity must be >= 0")
	}
	if l.MaxCapacity < l.MinCapacity {
		return errors.New("maxCapacity must be >= minCapacity")
	}
    switch l.Backpressure {
    case "", BackpressureBlock, BackpressureDropOldest, BackpressureDropNewest:
    default:
        return errors.New("invalid backpressure policy")
    }
	return nil
}

// Count returns the number of queued events.
// A nil receiver is treated as empty and returns 0.
func (l *LIFO) Count() uint32 {
	if l == nil {
		return 0
	}
	return atomic.LoadUint32(&l.n)
}

// Push appends an event to the top of the stack. When bounded and full:
//   - block:       returns ErrFull
//   - dropOldest:  drops the oldest (bottom) to keep the most recent items
//   - dropNewest:  drops the incoming item (no enqueue)
func (l *LIFO) Push(e Event) error {
    l.mu.Lock()
    defer l.mu.Unlock()
    if l.MaxCapacity > 0 && len(l.q) >= l.MaxCapacity {
        switch l.Backpressure {
        case BackpressureBlock:
            return ErrFull
        case BackpressureDropOldest:
            if len(l.q) > 0 {
                l.q = l.q[1:]
            } // drop oldest (bottom)
        case BackpressureDropNewest:
            // drop incoming; nothing enqueued
            return nil
        }
    } else {
        atomic.AddUint32(&l.n, 1)
    }
    l.q = append(l.q, any(e))
    return nil
}

// Pop removes and returns the most recent event. ok=false when empty.
func (l *LIFO) Pop() (Event, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := len(l.q)
	if n == 0 {
		var z Event
		return z, false
	}
	v := l.q[n-1]
	l.q = l.q[:n-1]
	atomic.AddUint32(&l.n, ^uint32(0)) // -1
	if ev, ok := v.(Event); ok {
		return ev, true
	}
	var z Event
	return z, false
}
