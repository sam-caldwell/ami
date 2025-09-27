package edge

import "fmt"

// FIFO models a first-in, first-out queue edge with optional bounds and backpressure.
type FIFO struct {
    MinCapacity   int    // minimum capacity (>=0)
    MaxCapacity   int    // maximum capacity (0 for unbounded; Max>=Min when nonzero)
    Backpressure  string // one of: block, dropOldest, dropNewest, shuntNewest, shuntOldest
}

func (f FIFO) Kind() Kind { return KindFIFO }

// Bounded reports whether a positive MaxCapacity was set.
func (f FIFO) Bounded() bool { return f.MaxCapacity > 0 }

// Delivery derives a human label for delivery semantics used in debug outputs.
// block -> atLeastOnce; others -> bestEffort.
func (f FIFO) Delivery() string {
    if f.Backpressure == "block" { return "atLeastOnce" }
    return "bestEffort"
}

func (f FIFO) Validate() error {
    if f.MinCapacity < 0 { return fmt.Errorf("minCapacity must be >= 0") }
    if f.MaxCapacity < 0 { return fmt.Errorf("maxCapacity must be >= 0 (or 0 for unbounded)") }
    if f.MaxCapacity > 0 && f.MaxCapacity < f.MinCapacity {
        return fmt.Errorf("maxCapacity must be >= minCapacity when bounded")
    }
    switch f.Backpressure {
    case "", "block", "dropOldest", "dropNewest", "shuntNewest", "shuntOldest":
        return nil
    default:
        return fmt.Errorf("invalid backpressure: %q", f.Backpressure)
    }
}

