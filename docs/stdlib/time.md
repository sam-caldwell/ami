# Stdlib: time

The `time` stdlib package provides a minimal wall-clock API: sleep, current time, arithmetic, and timestamps.

API (Go package `amitime`)
- `type Duration` (alias to Go's `time.Duration`).
- `type Time` (wraps Go's `time.Time`).
- `Now() Time`: current time.
- `Sleep(d Duration)`: pause for at least `d`.
- `Add(t Time, d Duration) Time`: `t + d`.
- `Delta(t1, t2 Time) Duration`: `t2 - t1`.
- `FromUnix(sec, nsec int64) Time`: construct from Unix epoch.
- `Time.Unix() int64`: seconds since epoch.
- `Time.UnixNano() int64`: nanoseconds since epoch.
// Ticker
- `type Ticker`
- `NewTicker(d Duration) *Ticker`: construct a ticker for period `d`.
- `(*Ticker).Register(fn func())`: add a handler to run every tick.
- `(*Ticker).Start()`: start ticking.
- `(*Ticker).Stop()`: stop ticking.

Notes
- `Delta(t1, t2)` returns a positive duration when `t2` occurs after `t1`.
- `Sleep` uses Go's monotonic clock where available.

Examples
```go
start := amitime.Now()
amitime.Sleep(100 * time.Millisecond)
end := amitime.Now()
if amitime.Delta(start, end) < (100 * time.Millisecond) {
  // handle clock skew / scheduling
}
```

Tests
- See `src/ami/stdlib/time/time_test.go` for basic behavior checks.
 - See `src/ami/stdlib/time/ticker_test.go` for ticker behavior.
