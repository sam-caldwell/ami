package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
)

// Verify human output includes parse errors, missingInSum, and unsatisfied when present.
func TestModAudit_Run_Human_ErrorsAndMissing(t *testing.T) {
    dir := filepath.Join("build", "test", "cli_mod_audit", "human_errors")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // workspace with imports that trigger parse errors and unsatisfied/missing
    ws := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: [ 'modA ^1.2.0', 'modB >= 1.0.0', 'modC 1.2.3', 'modD <= 1.2.3' ]\n")
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), ws, 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    // minimal ami.sum with modA ok, modB too old
    sum := []byte("{\n  \"schema\": \"ami.sum/v1\",\n  \"packages\": {\n    \"modA\": {\n      \"1.2.3\": \"shaA\"\n    },\n    \"modB\": {\n      \"0.9.0\": \"shaB\"\n    }\n  }\n}\n")
    if err := os.WriteFile(filepath.Join(dir, "ami.sum"), sum, 0o644); err != nil { t.Fatalf("write sum: %v", err) }

    var out bytes.Buffer
    if err := runModAudit(&out, dir, false); err != nil { t.Fatalf("run: %v", err) }
    s := out.String()
    if !containsAll(s, []string{"parse errors:", "missing in sum: modC", "unsatisfied: modB"}) {
        t.Fatalf("expected summary to include parse errors, missingInSum, unsatisfied; out=%s", s)
    }
}

func containsAll(hay string, needles []string) bool {
    for _, n := range needles {
        if !bytes.Contains([]byte(hay), []byte(n)) { return false }
    }
    return true
}
