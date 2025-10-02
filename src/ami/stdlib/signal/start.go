package amsignal

import "os"

// start initializes the shared signal channel and dispatcher loop.
func start() {
    ch = make(chan os.Signal, 4)
    go func() {
        for s := range ch {
            // Map os.Signal back to our enum set
            st := fromOSSignal(s)
            if st == 0 { continue }
            mu.Lock()
            fns := append([]func(){}, handlers[st]...)
            mu.Unlock()
            for _, f := range fns { safeCall(f) }
        }
    }()
}

