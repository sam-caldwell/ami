package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify that when per-env objects exist under build/<env>/obj/**, the verbose plan includes them.
func TestRunBuild_Verbose_PlanIncludesEnvObjects(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "plan_env")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    // Force a non-default env to ensure isolation from host default
    ws.Toolchain.Compiler.Env = []string{"linux/amd64"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    // Precreate a per-env object and index so the plan can discover it
    envObj := filepath.Join(dir, "build", "linux/amd64", "obj", ws.Packages[0].Package.Name)
    if err := os.MkdirAll(envObj, 0o755); err != nil { t.Fatalf("mkdir env obj: %v", err) }
    // object file
    if err := os.WriteFile(filepath.Join(envObj, "unit.o"), []byte{0}, 0o644); err != nil { t.Fatalf("write .o: %v", err) }
    // index file
    if err := os.WriteFile(filepath.Join(envObj, "index.json"), []byte(`{"schema":"objindex.v1","package":"`+ws.Packages[0].Package.Name+`","units":[]}`), 0o644); err != nil { t.Fatalf("write index: %v", err) }

    if err := runBuild(os.Stdout, dir, false, true); err != nil { t.Fatalf("runBuild: %v", err) }
    // Read plan
    p := filepath.Join(dir, "build", "debug", "build.plan.json")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read: %v", err) }
    var plan map[string]any
    if e := json.Unmarshal(b, &plan); e != nil { t.Fatalf("json: %v; %s", e, string(b)) }
    obe, _ := plan["objectsByEnv"].(map[string]any)
    if obe == nil || obe["linux/amd64"] == nil { t.Fatalf("objectsByEnv missing linux/amd64: %v", plan) }
    oie, _ := plan["objIndexByEnv"].(map[string]any)
    if oie == nil || oie["linux/amd64"] == nil { t.Fatalf("objIndexByEnv missing linux/amd64: %v", plan) }
}

