package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// In verbose mode, build writes kvstore metrics and dump under build/debug/kv/.
func TestRunBuild_Verbose_EmitsKVStoreMetricsAndDump(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "verbose_kv")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := runBuild(os.Stdout, dir, false, true); err != nil { t.Fatalf("runBuild: %v", err) }
    // metrics exists and has schema kv.metrics.v1
    mb, err := os.ReadFile(filepath.Join(dir, "build", "debug", "kv", "metrics.json"))
    if err != nil { t.Fatalf("metrics missing: %v", err) }
    var m map[string]any
    if e := json.Unmarshal(mb, &m); e != nil { t.Fatalf("json: %v; %s", e, string(mb)) }
    if sc, _ := m["schema"].(string); sc != "kv.metrics.v1" { t.Fatalf("schema: %v", sc) }
    // dump exists and has schema kv.dump.v1
    db, err := os.ReadFile(filepath.Join(dir, "build", "debug", "kv", "dump.json"))
    if err != nil { t.Fatalf("dump missing: %v", err) }
    var d map[string]any
    if e := json.Unmarshal(db, &d); e != nil { t.Fatalf("json: %v; %s", e, string(db)) }
    if sc, _ := d["schema"].(string); sc != "kv.dump.v1" { t.Fatalf("schema: %v", sc) }
}

