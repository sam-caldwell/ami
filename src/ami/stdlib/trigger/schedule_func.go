package trigger

import (
    stdtime "time"
    amitime "github.com/sam-caldwell/ami/src/ami/stdlib/time"
)

// Schedule emits a single Event[amitime.Time] at the specified time, then stops.
func Schedule(at amitime.Time) (<-chan Event[amitime.Time], func()) {
    out := make(chan Event[amitime.Time], 1)
    now := stdtime.Now()
    d := at.UnixNano() - now.UnixNano()
    if d < 0 { d = 0 }
    timer := stdtime.NewTimer(stdtime.Duration(d))
    done := make(chan struct{})
    go func() {
        defer close(out)
        select {
        case tm := <-timer.C:
            out <- Event[amitime.Time]{Value: toAMI(tm), Timestamp: amitime.Now()}
        case <-done:
            if !timer.Stop() { <-timer.C }
        }
    }()
    stop := func() { close(done) }
    return out, stop
}

