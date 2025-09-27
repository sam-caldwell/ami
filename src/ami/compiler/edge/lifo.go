package edge

import "fmt"

// LIFO models a last-in, first-out stack edge with optional bounds and backpressure.
type LIFO struct {
    MinCapacity   int
    MaxCapacity   int
    Backpressure  string // one of: block, dropOldest, dropNewest, shuntNewest, shuntOldest
}

func (l LIFO) Kind() Kind { return KindLIFO }

func (l LIFO) Bounded() bool { return l.MaxCapacity > 0 }

func (l LIFO) Delivery() string {
    if l.Backpressure == "block" { return "atLeastOnce" }
    return "bestEffort"
}

func (l LIFO) Validate() error {
    if l.MinCapacity < 0 { return fmt.Errorf("minCapacity must be >= 0") }
    if l.MaxCapacity < 0 { return fmt.Errorf("maxCapacity must be >= 0 (or 0 for unbounded)") }
    if l.MaxCapacity > 0 && l.MaxCapacity < l.MinCapacity {
        return fmt.Errorf("maxCapacity must be >= minCapacity when bounded")
    }
    switch l.Backpressure {
    case "", "block", "dropOldest", "dropNewest", "shuntNewest", "shuntOldest":
        return nil
    default:
        return fmt.Errorf("invalid backpressure: %q", l.Backpressure)
    }
}

