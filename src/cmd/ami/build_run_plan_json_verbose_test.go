package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Validate that the build plan is written when --json and --verbose are both set.
func TestRunBuild_Verbose_JSON_WritesBuildPlan(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "plan_json_verbose")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }

    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    if err := runBuild(os.Stdout, dir, true, true); err != nil {
        t.Fatalf("runBuild: %v", err)
    }
    p := filepath.Join(dir, "build", "debug", "build.plan.json")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read plan: %v", err) }
    var m map[string]any
    if e := json.Unmarshal(b, &m); e != nil { t.Fatalf("json: %v; %s", e, string(b)) }
    if m["schema"] != "build.plan/v1" { t.Fatalf("schema: %v", m["schema"]) }
}
