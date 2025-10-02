package buffer

import "sync"

// LIFOStack implements a bounded LIFO stack with backpressure.
type LIFOStack struct {
    MinCapacity int
    MaxCapacity int
    Backpressure string

    mu sync.Mutex
    s  []any
    pushN int
    popN  int
    dropN int
    fullN int
}

// Constructor moved to lifo_new.go to satisfy single-declaration rule

func (l *LIFOStack) Push(v any) error {
    l.mu.Lock(); defer l.mu.Unlock()
    l.pushN++
    cap := l.MaxCapacity
    if cap > 0 && len(l.s) >= cap {
        switch l.Backpressure {
        case "block":
            l.fullN++
            return ErrFull
        case "dropOldest", "shuntOldest":
            if len(l.s) > 0 {
                copy(l.s[0:], l.s[1:])
                l.s = l.s[:len(l.s)-1]
                l.dropN++
            }
        case "dropNewest", "shuntNewest":
            l.dropN++
            return nil
        default:
            l.dropN++
            return nil
        }
    }
    l.s = append(l.s, v)
    return nil
}

func (l *LIFOStack) Pop() (any, bool) {
    l.mu.Lock(); defer l.mu.Unlock()
    if len(l.s) == 0 { return nil, false }
    i := len(l.s)-1
    v := l.s[i]
    l.s = l.s[:i]
    l.popN++
    return v, true
}

func (l *LIFOStack) Len() int { l.mu.Lock(); defer l.mu.Unlock(); return len(l.s) }
func (l *LIFOStack) Counters() (push, pop, drop, full int) { l.mu.Lock(); defer l.mu.Unlock(); return l.pushN, l.popN, l.dropN, l.fullN }
