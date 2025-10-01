package time

import (
    stdtime "time"
    "sync"
)

// Duration is a Go-compatible duration type (nanoseconds).
type Duration = stdtime.Duration

// Time represents a time instant.
type Time = stdtime.Time

// Sleep pauses the current goroutine for at least the duration d.
func Sleep(d Duration) { stdtime.Sleep(d) }

// Now returns the current local time.
func Now() Time { return stdtime.Now() }

// Delta returns the duration t2 - t1.
func Delta(t1, t2 Time) Duration { return t2.Sub(t1) }

// Add returns the time t advanced by duration d.
func Add(t Time, d Duration) Time { return t.Add(d) }

// Ticker delivers events at intervals.
type Ticker struct {
    t    *stdtime.Ticker
    fn   func()
    quit chan struct{}
    stopOnce sync.Once
}

// NewTicker returns a new Ticker for duration d.
func NewTicker(d Duration) *Ticker {
    return &Ticker{t: stdtime.NewTicker(d), quit: make(chan struct{})}
}

// Register sets the function to be executed on each tick.
func (tk *Ticker) Register(f func()) { tk.fn = f }

// Start begins executing the registered function every tick.
func (tk *Ticker) Start() {
    if tk == nil || tk.t == nil { return }
    go func() {
        for {
            select {
            case <-tk.t.C:
                if tk.fn != nil { tk.fn() }
            case <-tk.quit:
                return
            }
        }
    }()
}

// Stop stops the ticker.
func (tk *Ticker) Stop() {
    if tk == nil || tk.t == nil { return }
    tk.stopOnce.Do(func(){ tk.t.Stop(); close(tk.quit) })
}
