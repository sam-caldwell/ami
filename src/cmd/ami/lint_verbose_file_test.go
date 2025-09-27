package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunLint_Verbose_WritesDebugFile(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "verbose")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, true, false); err != nil { t.Fatalf("runLint: %v", err) }
    // Verify debug file exists and contains schema marker
    p := filepath.Join(dir, "build", "debug", "lint.ndjson")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read debug file: %v", err) }
    if !bytes.Contains(b, []byte("\"schema\":\"diag.v1\"")) {
        t.Fatalf("expected diag.v1 records, got: %s", string(b))
    }
}
