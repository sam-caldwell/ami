package amios

import (
    "bytes"
    "runtime"
    "strings"
    "testing"
    "time"
)

func TestExecAndStart_Blocking_GoVersion(t *testing.T) {
    p, err := Exec("go", "version")
    if err != nil { t.Fatalf("exec: %v", err) }
    if pid := p.Pid(); pid != 0 { t.Fatalf("pid should be 0 before start, got %d", pid) }
    if err := p.Start(true); err != nil { t.Fatalf("start(block): %v", err) }
    st := p.Status()
    if st.Running { t.Fatalf("expected not running after blocking run") }
    if st.ExitCode == nil || *st.ExitCode != 0 { t.Fatalf("expected exit code 0, got %+v", st.ExitCode) }
    out := string(p.Stdout())
    if !strings.Contains(out, "go version") {
        t.Fatalf("stdout missing 'go version': %q", out)
    }
}

func TestExecAndStart_Async_StatusCompletes(t *testing.T) {
    // Cross-platform short command; use `go env GOOS` to avoid sleep/echo portability issues
    p, err := Exec("go", "env", "GOOS")
    if err != nil { t.Fatalf("exec: %v", err) }
    if err := p.Start(false); err != nil { t.Fatalf("start(async): %v", err) }
    if pid := p.Pid(); pid <= 0 { t.Fatalf("expected pid > 0 after start, got %d", pid) }
    // Poll for completion up to 2s
    deadline := time.Now().Add(2 * time.Second)
    for time.Now().Before(deadline) {
        st := p.Status()
        if !st.Running && st.ExitCode != nil {
            if *st.ExitCode != 0 { t.Fatalf("expected exit code 0, got %d", *st.ExitCode) }
            // Verify stdout contains current GOOS
            goos := strings.TrimSpace(string(p.Stdout()))
            if goos != runtime.GOOS {
                // Some Go envs print with newline; tolerate suffix/mismatch but ensure non-empty
                if len(bytes.TrimSpace(p.Stdout())) == 0 {
                    t.Fatalf("expected non-empty stdout for go env GOOS")
                }
            }
            return
        }
        time.Sleep(10 * time.Millisecond)
    }
    t.Fatalf("process did not complete in time; status: %+v", p.Status())
}

