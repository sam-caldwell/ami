package amsignal

import (
    gosignal "os/signal"
    "sync"
)

// Reset clears all registered handlers and stops notifications (for tests).
func Reset() {
    mu.Lock()
    defer mu.Unlock()
    for st := range handlers { handlers[st] = nil }
    if ch != nil {
        gosignal.Stop(ch)
        close(ch)
        ch = nil
    }
    // allow re-init
    once = sync.Once{}
}
