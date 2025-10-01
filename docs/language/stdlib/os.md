# Stdlib: os (Process Runner)

The `os` stdlib package exposes a minimal, cross‑platform process runner and basic system/environment helpers.

API (Go package `amios`)
- `Exec(program string, args ...string) (*Process, error)`: construct a process without starting it.
- `(*Process).Start(block bool) error`: start; when `block=true`, run to completion and record exit code; when `false`, return immediately and track exit asynchronously.
- `(*Process).Kill() error`: best‑effort immediate termination. May vary across platforms.
- `(*Process).Pid() int`: OS process id after `Start`; `0` before start.
- `(*Process).Status() ProcessStatus`: snapshot `{ PID int, Running bool, ExitCode *int }`.
- `(*Process).Stdin() io.WriteCloser`: a pipe writer to the child’s stdin, usable before or after `Start()`.
- `(*Process).Stdout() []byte`, `(*Process).Stderr() []byte`: bytes captured so far from child stdout/stderr.
- `GetSystemStats() SystemStats`: `{ OS, Arch, NumCPU, TotalMemoryBytes }` (best‑effort memory, see notes).
- Env helpers: `GetEnv(name string) string`, `SetEnv(name, value string) error`, `ListEnv() []string` (names only).

Notes
- `Stdin()` is pre‑wired via `io.Pipe()` at `Exec()` time so input can be written before or after `Start()`.
- `Kill()` is validated on Linux in tests; other platforms rely on `os/exec` best‑effort semantics.
- `GetSystemStats.TotalMemoryBytes` is populated on Linux (`/proc/meminfo`) and macOS (`sysctl -n hw.memsize`); other platforms may return `0`.
- `Status()` is race‑free for inspection and returns a copy with an optional `ExitCode` pointer when known.

Examples
- Echo stdin to stdout and capture output
  ```go
  p, _ := amios.Exec("go", "run", "echo.go")
  _ = p.Start(false)
  w, _ := p.Stdin()
  _, _ = w.Write([]byte("hello\n"))
  _ = w.Close()
  // poll until exit, then:
  out := string(p.Stdout()) // "hello\n"
  ```
- Get system stats
  ```go
  st := amios.GetSystemStats()
  _ = st.OS; _ = st.Arch; _ = st.NumCPU; _ = st.TotalMemoryBytes
  ```
- Environment helpers
  ```go
  _ = amios.SetEnv("AMI_EXAMPLE", "1")
  v := amios.GetEnv("AMI_EXAMPLE") // "1"
  names := amios.ListEnv()          // ["PATH", "HOME", ...]
  _ = v; _ = names
  ```

