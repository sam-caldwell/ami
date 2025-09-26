package edge

import (
	"errors"
	"sync"
	"sync/atomic"
)

// Pipeline is a declarative, compiler-generated bridge between pipelines.
// It identifies the upstream pipeline by its ingress name and declares the
// capacity/backpressure policy of the bridging queue.
type Pipeline struct {
	UpstreamName string
	MinCapacity  int
	MaxCapacity  int
	Backpressure BackpressurePolicy
	TypeName     string // payload type name for analysis/codegen
	// runtime buffer scaffold (FIFO semantics)
	mu sync.Mutex
	q  []any
	n  uint32
}

func (p *Pipeline) Kind() string { return "edge.Pipeline" }

func (p *Pipeline) Validate() error {
	if p.UpstreamName == "" {
		return errors.New("upstream pipeline name required")
	}
	if p.MinCapacity < 0 {
		return errors.New("minCapacity must be >= 0")
	}
	if p.MaxCapacity < p.MinCapacity {
		return errors.New("maxCapacity must be >= minCapacity")
	}
    switch p.Backpressure {
    case "", BackpressureBlock, BackpressureDropOldest, BackpressureDropNewest:
    default:
        return errors.New("invalid backpressure policy")
    }
	return nil
}

// Count returns the number of queued events.
// A nil receiver is treated as empty and returns 0.
func (p *Pipeline) Count() uint32 {
	if p == nil {
		return 0
	}
	return atomic.LoadUint32(&p.n)
}

// Push enqueues an event onto the bridging queue (FIFO semantics). When bounded
// and full:
//   - block:       ErrFull
//   - dropOldest:  drop oldest
//   - dropNewest:  drop incoming
func (p *Pipeline) Push(e Event) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    if p.MaxCapacity > 0 && len(p.q) >= p.MaxCapacity {
        switch p.Backpressure {
        case BackpressureBlock:
            return ErrFull
        case BackpressureDropOldest:
            if len(p.q) > 0 {
                p.q = p.q[1:]
            } // drop oldest
        case BackpressureDropNewest:
            // drop incoming
            return nil
        }
    } else {
        atomic.AddUint32(&p.n, 1)
    }
    p.q = append(p.q, any(e))
    return nil
}

// Pop dequeues the oldest event. ok=false when empty.
func (p *Pipeline) Pop() (Event, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.q) == 0 {
		var z Event
		return z, false
	}
	v := p.q[0]
	p.q = p.q[1:]
	atomic.AddUint32(&p.n, ^uint32(0)) // -1
	if ev, ok := v.(Event); ok {
		return ev, true
	}
	var z Event
	return z, false
}
