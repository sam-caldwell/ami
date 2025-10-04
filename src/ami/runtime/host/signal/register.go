package amsignal

import (
    "os"
    gosignal "os/signal"
    "sync"
)

var (
    mu       sync.Mutex
    once     sync.Once
    ch       chan os.Signal
    handlers = map[SignalType][]func(){}
)

// Register installs fn as a handler for the given signal. Multiple handlers may be
// registered per signal. Handlers are invoked sequentially when a signal arrives.
func Register(sig SignalType, fn func()) {
    mu.Lock()
    defer mu.Unlock()
    handlers[sig] = append(handlers[sig], fn)
    // ensure goroutine started once; subscribe for signals we know about
    once.Do(start)
    gosignal.Notify(ch, toOSSignal(sig))
}

