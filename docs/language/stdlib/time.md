# Stdlib: time (Sleep, Timestamps, Ticker)

The `time` stdlib package wraps common time utilities with simple, testable APIs.

API (Go package `time`)
- `type Duration = time.Duration`: alias to Go duration (nanoseconds).
- `type Time = time.Time`: alias to Go time instant.
- `Sleep(d Duration)`: pause for at least `d`.
- `Now() Time`: current local time.
- `Delta(t1, t2 Time) Duration`: compute `t2 - t1`.
- `Add(t Time, d Duration) Time`: `t + d`.
- `type Ticker`: periodic callback executor.
  - `NewTicker(d Duration) *Ticker`: create ticker with period `d`.
  - `(*Ticker).Register(func())`: set the function to execute on each tick.
  - `(*Ticker).Start()`: start background tick loop.
  - `(*Ticker).Stop()`: stop ticker and exit loop.

Notes
- `Ticker` spawns a goroutine in `Start()`; call `Stop()` to release resources.
- `Now().Unix()` and `Now().UnixNano()` remain available on the alias type.

Examples
```go
import stdtime "time" // Go stdlib alias to avoid name clash

t1 := time.Now()
time.Sleep(10 * stdtime.Millisecond)
t2 := time.Now()
dt := time.Delta(t1, t2)
_ = dt

count := 0
tk := time.NewTicker(5 * stdtime.Millisecond)
tk.Register(func(){ count++ })
tk.Start()
time.Sleep(20 * stdtime.Millisecond)
tk.Stop()
```
