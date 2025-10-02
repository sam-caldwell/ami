package main

import (
    "bytes"
    "encoding/json"
    "strings"
    "testing"
)

func Test_modGetGit_fileGitRequiresAbsolutePath(t *testing.T) {
    dir := t.TempDir()
    var buf bytes.Buffer
    err := modGetGit(&buf, dir, "file+git://relative/path#v1.0.0", true)
    if err == nil { t.Fatal("expected error for relative file+git path") }
    var r modGetResult
    if decErr := json.Unmarshal(buf.Bytes(), &r); decErr != nil {
        t.Fatalf("json decode failed: %v (out=%q)", decErr, buf.String())
    }
    if !strings.Contains(strings.ToLower(r.Message), "file+git") {
        t.Fatalf("unexpected message: %q", r.Message)
    }
}

