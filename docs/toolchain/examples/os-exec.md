# Example (Planned): Using stdlib os from AMI

This is a forward-looking example showing how AMI code could use the stdlib `os` package once language hooks are available.

AMI sample (illustrative only):

```
package app
import os

// Run a short external program and return its PID when successful.
func RunEcho() (int) {
  var p os.Process
  p = os.Exec("echo", "hello")
  p.Start(true) // block until finished
  var st = p.Status()
  if st.ExitCode != 0 { return -1 }
  return p.Pid()
}

pipeline P(){
  ingress;
  // Optionally call RunEcho() from a worker to integrate with a pipeline.
  egress;
}
```

Notes
- The `os` API in AMI provides `Exec`, `Process.Start/Kill/Status/Stdin/Stdout/Stderr/Pid`, and `GetSystemStats()`.
- This example will compile and run once the language/runtime hooks for stdlib packages are enabled.
- For today, see `docs/language/stdlib/os.md` and the Go tests under `src/ami/stdlib/os/*_test.go` for working behavior.
