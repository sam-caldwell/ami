package trigger

import (
    stdtime "time"
    amitime "github.com/sam-caldwell/ami/src/ami/runtime/host/time"
)

// Timer emits an Event[amitime.Time] every d interval until stop is invoked.
func Timer(d amitime.Duration) (<-chan Event[amitime.Time], func()) {
    out := make(chan Event[amitime.Time], 16)
    t := stdtime.NewTicker(stdtime.Duration(d))
    done := make(chan struct{})
    go func() {
        for {
            select {
            case tm := <-t.C:
                out <- Event[amitime.Time]{Value: toAMI(tm), Timestamp: amitime.Now()}
            case <-done:
                t.Stop()
                close(out)
                return
            }
        }
    }()
    stop := func() { close(done) }
    return out, stop
}

