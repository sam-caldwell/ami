package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_JSON_Verbose_SummaryIncludesObjIndex(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "json_verbose_objindex")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var out bytes.Buffer
    if err := runBuild(&out, dir, true, true); err != nil { t.Fatalf("runBuild: %v", err) }
    var m map[string]any
    if e := json.Unmarshal(out.Bytes(), &m); e != nil { t.Fatalf("json: %v; %s", e, out.String()) }
    arr, ok := m["data"].(map[string]any)["objIndex"].([]any)
    if !ok || len(arr) == 0 { t.Fatalf("objIndex missing in summary: %v", m) }
}

