package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_Success_JSONAndHuman(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "ok")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    // No remote imports ensures audit is a no-op
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    // JSON mode
    var jb bytes.Buffer
    if err := runBuild(&jb, dir, true); err != nil { t.Fatalf("runBuild json: %v", err) }
    var m map[string]any
    if e := json.Unmarshal(jb.Bytes(), &m); e != nil { t.Fatalf("json: %v; out=%s", e, jb.String()) }
    if m["code"] != "BUILD_OK" { t.Fatalf("expected BUILD_OK; got %v", m["code"]) }

    // Human mode
    var hb bytes.Buffer
    if err := runBuild(&hb, dir, false); err != nil { t.Fatalf("runBuild human: %v", err) }
    if hb.Len() == 0 { t.Fatalf("expected human message output") }
}
