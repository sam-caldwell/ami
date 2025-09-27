package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verifies that mapping a warning code to error via workspace rules causes
// a non-zero exit in JSON mode (errorsN > 0) and the level is elevated.
func TestLint_RuleMapping_ElevateToError_JSONExit(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "rule_elevate")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    // Create unsorted imports to trigger W_IMPORT_ORDER
    ws.Packages[0].Package.Import = []string{"zeta", "alpha"}
    if ws.Toolchain.Linter.Rules == nil { ws.Toolchain.Linter.Rules = map[string]string{} }
    ws.Toolchain.Linter.Rules["W_IMPORT_ORDER"] = "error"
    ws.Toolchain.Linter.Options = []string{} // ensure non-strict; rules mapping drives elevation
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
        t.Fatalf("save: %v", err)
    }
    var buf bytes.Buffer
    err := runLint(&buf, dir, true, false, false)
    if err == nil {
        t.Fatalf("expected non-nil error due to elevated rule causing errorsN>0 in JSON mode")
    }
    // Verify the record shows level:error for W_IMPORT_ORDER
    dec := json.NewDecoder(&buf)
    var sawElevated bool
    for dec.More() {
        var m map[string]any
        if e := dec.Decode(&m); e != nil { t.Fatalf("json: %v", e) }
        if m["code"] == "W_IMPORT_ORDER" {
            if m["level"] != "error" { t.Fatalf("expected elevated level 'error', got %v", m["level"]) }
            sawElevated = true
        }
    }
    if !sawElevated { t.Fatalf("expected W_IMPORT_ORDER record in output: %s", buf.String()) }
}
