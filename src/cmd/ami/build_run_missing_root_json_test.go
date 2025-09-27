package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_JSON_MissingPackageRoot_EmitsDiagIO(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "json_missing_root")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    // set non-existent root
    ws.Packages[0].Package.Root = "./does_not_exist"
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var out bytes.Buffer
    err := runBuild(&out, dir, true, true)
    if err == nil { t.Fatalf("expected error") }
    var rec map[string]any
    if e := json.Unmarshal(out.Bytes(), &rec); e != nil { t.Fatalf("json: %v; out=%s", e, out.String()) }
    if rec["schema"] != "diag.v1" || rec["code"] != "E_FS_MISSING" { t.Fatalf("diag: %+v", rec) }
}

