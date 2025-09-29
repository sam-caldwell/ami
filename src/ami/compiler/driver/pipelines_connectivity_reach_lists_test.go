package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestPipelinesDebug_Connectivity_UnreachableLists(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // Only edge: ingress -> A. B and egress are unreachable from ingress.
    // A cannot reach egress (no path to egress).
    code := "package app\npipeline P(){ ingress; A(); B(); egress; ingress -> A; }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    p := filepath.Join("build", "debug", "ir", "app", "u.pipelines.json")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    pipes := obj["pipelines"].([]any)
    ent := pipes[0].(map[string]any)
    conn := ent["connectivity"].(map[string]any)
    ufi := toStrings(conn["unreachableFromIngress"])
    cre := toStrings(conn["cannotReachEgress"])
    // Expect at least B and egress unreachable from ingress
    if !hasStr(ufi, "B") { t.Fatalf("unreachableFromIngress missing B: %v", ufi) }
    if !hasStr(ufi, "egress") { t.Fatalf("unreachableFromIngress missing egress: %v", ufi) }
    // Expect A cannot reach egress
    if !hasStr(cre, "A") { t.Fatalf("cannotReachEgress missing A: %v", cre) }
}

func toStrings(v any) []string {
    arr, ok := v.([]any)
    if !ok { return nil }
    out := make([]string, 0, len(arr))
    for _, e := range arr { if s, ok := e.(string); ok { out = append(out, s) } }
    return out
}

func hasStr(ss []string, want string) bool {
    for _, s := range ss { if s == want { return true } }
    return false
}

