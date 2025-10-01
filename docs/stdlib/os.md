# Stdlib: os (Process Runner)

The `os` stdlib package provides a minimal process runner and system information.

API (Go package `amios`)
- `Exec(program string, args ...string) (*Process, error)`: initialize a process without starting it.
- `(*Process).Start(block bool) error`: start the process; when `block=true`, waits for completion.
- `(*Process).Kill() error`: terminate the process immediately (best-effort across platforms).
- `(*Process).Status() ProcessStatus`: returns `{ PID, Running, ExitCode* }`.
- `(*Process).Pid() int`: OS process id (0 before Start).
- `(*Process).Stdin() io.WriteCloser`: returns a writer to process stdin.
- `(*Process).Stdout() []byte`, `(*Process).Stderr() []byte`: captured outputs so far.
- `GetSystemStats() SystemStats`: returns `{ OS, Arch, NumCPU, TotalMemoryBytes }` (best-effort memory).

Notes
- `Stdin()` is available before or after `Start()`; the implementation wires a pipe at `Exec()` time.
- `Kill()` semantics are platform dependent; tests validate kill behavior on Linux.
- Total memory is populated on Linux (`/proc/meminfo`) and macOS (`sysctl -n hw.memsize`); other platforms return 0.

Examples
- Echo stdin to stdout and capture output
  ```go
  p, _ := amios.Exec("go", "run", "echo.go")
  _ = p.Start(false)
  w, _ := p.Stdin()
  w.Write([]byte("hello\n"))
  w.Close()
  // wait ... then
  out := string(p.Stdout()) // "hello\n"
  ```
- Get system stats
  ```go
  st := amios.GetSystemStats()
  // st.OS, st.Arch, st.NumCPU, st.TotalMemoryBytes
  ```
