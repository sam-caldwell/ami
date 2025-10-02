package main

import (
    "bytes"
    "encoding/json"
    "strings"
    "testing"
)

func Test_runModGet_missingWorkspace_JSON(t *testing.T) {
    dir := t.TempDir()
    var buf bytes.Buffer
    err := runModGet(&buf, dir, "some/src", true)
    if err == nil { t.Fatal("expected error when workspace is missing") }
    var r modGetResult
    if decErr := json.Unmarshal(buf.Bytes(), &r); decErr != nil {
        t.Fatalf("json decode failed: %v (out=%q)", decErr, buf.String())
    }
    if r.Source != "some/src" { t.Fatalf("unexpected source: %q", r.Source) }
    if !strings.Contains(r.Message, "workspace") { t.Fatalf("unexpected message: %q", r.Message) }
}

