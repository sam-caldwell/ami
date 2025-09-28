package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    es "github.com/sam-caldwell/ami/src/schemas/errors"
)

func TestErrorsValidate_HappyAndSad(t *testing.T) {
    dir := t.TempDir()
    good := filepath.Join(dir, "good.json")
    bad := filepath.Join(dir, "bad.json")
    // Good error record
    rec := es.Error{Level: "error", Code: "E_X", Message: "x"}
    if b, err := json.Marshal(rec); err != nil { t.Fatal(err) } else { os.WriteFile(good, b, 0o644) }
    // Bad record (missing code/message)
    if err := os.WriteFile(bad, []byte(`{"schema":"errors.v1","level":"error"}`), 0o644); err != nil { t.Fatal(err) }
    c := newRootCmd()
    c.SetArgs([]string{"errors", "validate", "--file", good})
    if err := c.Execute(); err != nil { t.Fatalf("good validate: %v", err) }
    c = newRootCmd()
    c.SetArgs([]string{"errors", "validate", "--file", bad})
    if err := c.Execute(); err == nil { t.Fatalf("expected error for bad error record") }
}

