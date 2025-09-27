package logging

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

type bufWriter struct{ b *strings.Builder }

func (w bufWriter) Write(p []byte) (int, error) { return w.b.WriteString(string(p)) }

func TestLogger_JSONIncludesTimestampAndNoANSI(t *testing.T) {
    var sb strings.Builder
    lg, err := New(Options{JSON: true, Out: bufWriter{&sb}})
    if err != nil {
        t.Fatalf("New logger: %v", err)
    }
    defer lg.Close()
    lg.Info("hello", map[string]any{"x": 1})
    out := sb.String()
    if !strings.Contains(out, `"timestamp":`) {
        t.Fatalf("missing timestamp field: %q", out)
    }
    if ansiPattern().MatchString(out) {
        t.Fatalf("JSON output should not contain ANSI codes: %q", out)
    }
}

func TestLogger_VerboseAlsoWritesDebugFile(t *testing.T) {
    // Place artifacts under repo-root/build/test/logging
    base := filepath.Join(repoRoot(t), "build", "test", "logging")
    if err := os.MkdirAll(base, 0o755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    var sb strings.Builder
    lg, err := New(Options{JSON: false, Verbose: true, DebugDir: base, Out: bufWriter{&sb}})
    if err != nil {
        t.Fatalf("New logger: %v", err)
    }
    lg.Info("debug file line", nil)
    _ = lg.Close()

    // Verify debug file contains the line
    data, err := os.ReadFile(filepath.Join(base, "activity.log"))
    if err != nil {
        t.Fatalf("read debug file: %v", err)
    }
    if !strings.Contains(string(data), "debug file line") {
        t.Fatalf("debug file missing content: %q", string(data))
    }
}

// repoRoot walks up from CWD to locate go.mod and returns that directory.
func repoRoot(t *testing.T) string {
    t.Helper()
    dir, err := os.Getwd()
    if err != nil {
        t.Fatalf("getwd: %v", err)
    }
    for i := 0; i < 10; i++ { // limit ascent
        if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
            return dir
        }
        parent := filepath.Dir(dir)
        if parent == dir { // reached volume root
            break
        }
        dir = parent
    }
    t.Fatalf("go.mod not found from test cwd")
    return ""
}
