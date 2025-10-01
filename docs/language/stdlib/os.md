# Stdlib: os (Process Runner)

The `os` module provides minimal, cross‑platform process execution and basic system/environment helpers for AMI programs.

API (AMI module `os`)
- `func os.exec(program string, args...string) (Process, error)` — construct a process without starting it.
- `method Process.start(block bool) error` — start; when `block=true`, wait for completion; when `false`, return immediately and track exit asynchronously.
- `method Process.kill() error` — best‑effort immediate termination.
- `method Process.pid() int` — OS process id after `start`; `0` before start.
- `method Process.status() (pid int, running bool, exitCode Optional<int>)` — snapshot of process state.
- `method Process.stdin() (Writer, error)` — write pipe to child stdin (usable before or after `start`).
- `method Process.stdout() bytes` — bytes captured so far from child stdout.
- `method Process.stderr() bytes` — bytes captured so far from child stderr.
- `func os.stats() (os string, arch string, numCPU int, totalMemoryBytes int64)` — best‑effort system info.
- `func os.getenv(name string) string`, `func os.setenv(name string, value string) error`, `func os.envnames() slice<string>` — environment helpers.

Notes
- `stdin()` is a pre‑wired pipe to the child, available before and after `start()`.
- Memory reporting in `os.stats()` is best‑effort and may return `0` on some platforms.
- Process control is subject to platform limits; treat `kill()` as best‑effort.

I/O Gating
- File and network access is governed by the `io` module’s capability policy. See docs/language/stdlib/io.md for details.

Examples (AMI)
- Echo stdin to stdout and capture output
  ```
  import os

  func RunEcho(){
    var p, _ = os.exec("/bin/cat")
    _ = p.start(false)
    var w, _ = p.stdin()
    w.write("hello\n")
    w.close()
    // poll until exit...
    var out = bytesToString(p.stdout()) // "hello\n"
  }
  ```
- Get system stats and env
  ```
  import os
  var (sys, arch, ncpu, mem) = os.stats()
  var _ = os.setenv("AMI_EXAMPLE", "1")
  var v = os.getenv("AMI_EXAMPLE")   // "1"
  var names = os.envnames()           // ["PATH", "HOME", ...]
  ```
