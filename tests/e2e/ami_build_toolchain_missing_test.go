package e2e

import (
    "bytes"
    "encoding/json"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

func TestE2E_AmiBuild_JSON_Reports_ToolchainMissing_WhenNoClang(t *testing.T) {
    // If clang is present, skip this test.
    if _, err := exec.LookPath("clang"); err == nil {
        t.Skip("clang present; skipping missing-toolchain test")
    }
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "build", "missing_toolchain")
    // Prepare minimal workspace
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n"), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.WriteFile(filepath.Join(ws, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write src: %v", err) }
    // Run `ami build --json` to surface diagnostics
    cmd := exec.Command(bin, "build", "--json")
    cmd.Dir = ws
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    if err := cmd.Run(); err == nil {
        t.Fatalf("expected non-zero due to diagnostics; stderr=%s\nstdout=%s", stderr.String(), stdout.String())
    }
    // Expect at least one diag record with code E_TOOLCHAIN_MISSING
    dec := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
    found := false
    for dec.More() {
        var rec map[string]any
        if err := dec.Decode(&rec); err != nil { break }
        if c, ok := rec["code"].(string); ok && c == "E_TOOLCHAIN_MISSING" { found = true; break }
    }
    if !found {
        t.Fatalf("expected E_TOOLCHAIN_MISSING in diagnostics; stdout=%s", stdout.String())
    }
}
