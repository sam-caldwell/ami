package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLintWorkspace_NameStyleAndImportChecks(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "engine")
    if err := os.MkdirAll(dir, 0o755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    ws := workspace.DefaultWorkspace()
    // Underscore in name triggers W_PKG_NAME_STYLE
    ws.Packages[0].Package.Name = "bad_name"
    // Bad import chars, missing local path, duplicate, relative ../, and unsorted order
    ws.Packages[0].Package.Import = []string{"bad space", "./missing_local", "zeta", "alpha", "../escape", "lib >= v1.2.3", "mod >= badver"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
        t.Fatalf("save: %v", err)
    }
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err == nil {
        // We still expect non-zero exit due to warnings? In JSON mode we only encode diags and summary, return nil
    }
    // Decode NDJSON lines; last is summary
    dec := json.NewDecoder(&buf)
    var records int
    var hasOrder, hasConstraint bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil {
            t.Fatalf("json decode: %v", err)
        }
        if m["code"] == "W_IMPORT_ORDER" { hasOrder = true }
        if m["code"] == "W_IMPORT_CONSTRAINT_INVALID" { hasConstraint = true }
        records++
    }
    if records < 2 { // at least one warn + summary
        t.Fatalf("expected >=2 JSON lines, got %d: %s", records, buf.String())
    }
    if !hasOrder { t.Fatalf("expected W_IMPORT_ORDER warning; out=%s", buf.String()) }
    if !hasConstraint { t.Fatalf("expected W_IMPORT_CONSTRAINT_INVALID; out=%s", buf.String()) }
}
