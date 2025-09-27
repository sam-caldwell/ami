package main

import (
    "bufio"
    "bytes"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_JSON_SyntaxDiagnostics_Multiple(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "json_syn_multi")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    // Construct a file with at least two parser errors
    bad := "pkg app\nfunc ( {\n}"
    if err := os.WriteFile(filepath.Join(dir, "src", "bad.ami"), []byte(bad), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var out bytes.Buffer
    _ = runBuild(&out, dir, true, false)
    // Count lines
    cnt := 0
    s := bufio.NewScanner(bytes.NewReader(out.Bytes()))
    for s.Scan() { cnt++ }
    if cnt < 2 { t.Fatalf("expected multiple diag lines; got %d\n%s", cnt, out.String()) }
}

