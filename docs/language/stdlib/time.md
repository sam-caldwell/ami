# Stdlib: time (Sleep, Timestamps, Ticker)

The `time` module exposes common time utilities to AMI code.

API (AMI module `time`)
- `type Duration` — duration value (nanoseconds resolution).
- `type Time` — time instant.
- `func time.sleep(d Duration)` — pause for at least `d`.
- `func time.now() Time` — current local time.
- `func time.delta(t1 Time, t2 Time) Duration` — `t2 - t1`.
- `func time.add(t Time, d Duration) Time` — `t + d`.
- `type Ticker` — periodic callback executor.
  - `func time.ticker(period Duration) Ticker` — create a ticker.
  - `method Ticker.onTick(f func())` — set the function to execute on each tick.
  - `method Ticker.start()` — start background tick loop.
  - `method Ticker.stop()` — stop ticker and exit loop.

Notes
- `Ticker` runs callbacks periodically once `start()` is called; call `stop()` to release resources.
- On `Time`, `unix()` and `unixNano()` are provided for timestamp extraction.

Examples (AMI)
```
import time

let t1 = time.now()
time.sleep(10ms)
let t2 = time.now()
let dt = time.delta(t1, t2)

var count = 0
let tk = time.ticker(5ms)
tk.onTick(func(){ count = count + 1 })
tk.start()
time.sleep(20ms)
tk.stop()
```
