package root_test

import (
    "bufio"
    "os"
    "strings"
    "testing"
)

// captureStdout captures stdout while fn executes and returns captured output as string.
func captureStdout(t *testing.T, fn func()) string {
    t.Helper()
    old := os.Stdout
    r, w, err := os.Pipe()
    if err != nil { t.Fatalf("pipe: %v", err) }
    os.Stdout = w
    defer func() { os.Stdout = old }()
    fn()
    _ = w.Close()
    var b strings.Builder
    sc := bufio.NewScanner(r)
    for sc.Scan() { b.WriteString(sc.Text()); b.WriteByte('\n') }
    return b.String()
}

