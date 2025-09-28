package logging

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"
)

// Test that environment variables can configure the pipeline (e.g., enable time-based flush).
func TestLogger_Pipeline_Config_FromEnv(t *testing.T) {
    base := filepath.Join(repoRoot(t), "build", "test", "logging_env")
    if err := os.MkdirAll(base, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }

    // Save and restore env
    oldFlush, hadFlush := os.LookupEnv("AMI_LOG_PIPE_FLUSH_INTERVAL")
    oldBatch, hadBatch := os.LookupEnv("AMI_LOG_PIPE_BATCH_MAX")
    t.Cleanup(func() {
        if hadFlush { _ = os.Setenv("AMI_LOG_PIPE_FLUSH_INTERVAL", oldFlush) } else { _ = os.Unsetenv("AMI_LOG_PIPE_FLUSH_INTERVAL") }
        if hadBatch { _ = os.Setenv("AMI_LOG_PIPE_BATCH_MAX", oldBatch) } else { _ = os.Unsetenv("AMI_LOG_PIPE_BATCH_MAX") }
    })
    _ = os.Setenv("AMI_LOG_PIPE_FLUSH_INTERVAL", "10ms")
    _ = os.Setenv("AMI_LOG_PIPE_BATCH_MAX", "10")

    var sb strings.Builder
    lg, err := New(Options{JSON: true, Verbose: true, DebugDir: base, Out: bufWriter{&sb}})
    if err != nil { t.Fatalf("New logger: %v", err) }

    // Write one line; do not close to ensure timer-based flush is required
    lg.Info("env configured flush", nil)
    time.Sleep(50 * time.Millisecond)

    data, err := os.ReadFile(filepath.Join(base, "activity.log"))
    if err != nil { t.Fatalf("read debug file: %v", err) }
    if !strings.Contains(string(data), "env configured flush") {
        t.Fatalf("expected time-based flush to write data, got: %q", string(data))
    }
    _ = lg.Close()
}

