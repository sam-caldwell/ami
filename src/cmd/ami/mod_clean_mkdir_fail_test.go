package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/exit"
)

func TestModClean_MkdirFails_WhenParentIsFile_JSON(t *testing.T) {
    base := filepath.Join("build", "test", "mod_clean", "mkdir_fail")
    _ = os.RemoveAll(base)
    if err := os.MkdirAll(base, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Create a file that will act as parent
    parent := filepath.Join(base, "parent")
    if err := os.WriteFile(parent, []byte("x"), 0o644); err != nil { t.Fatalf("write parent: %v", err) }
    // Target cache is under the file parent, which should fail on MkdirAll
    target := filepath.Join(parent, "child")
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", target)

    var buf bytes.Buffer
    err := runModClean(&buf, true)
    if err == nil { t.Fatalf("expected error when parent is file") }
    if exit.UnwrapCode(err) != exit.IO { t.Fatalf("expected exit.IO; got %v", exit.UnwrapCode(err)) }
    var res modCleanResult
    if e := json.Unmarshal(buf.Bytes(), &res); e != nil { t.Fatalf("json: %v; out=%s", e, buf.String()) }
    if res.Created { t.Fatalf("expected Created=false on mkdir failure") }
}

