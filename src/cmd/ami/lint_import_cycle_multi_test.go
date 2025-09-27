package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify multi-node cycle detection and that we emit exactly one canonicalized cycle.
func TestLint_CircularLocalImports_MultiNode_CanonicalizedAndDeduped(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "cycle_multi")
    aDir := filepath.Join(dir, "a")
    bDir := filepath.Join(dir, "b")
    cDir := filepath.Join(dir, "c")
    if err := os.MkdirAll(aDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.MkdirAll(bDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.MkdirAll(cDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    _ = os.WriteFile(filepath.Join(aDir, "main.ami"), []byte("package a\n"), 0o644)
    _ = os.WriteFile(filepath.Join(bDir, "main.ami"), []byte("package b\n"), 0o644)
    _ = os.WriteFile(filepath.Join(cDir, "main.ami"), []byte("package c\n"), 0o644)

    ws := workspace.DefaultWorkspace()
    ws.Toolchain.Linter.Options = []string{}
    ws.Packages = workspace.PackageList{
        {Key: "main", Package: workspace.Package{Name: "a", Version: "0.0.1", Root: "./a", Import: []string{"./b"}}},
        {Key: "b", Package: workspace.Package{Name: "b", Version: "0.0.1", Root: "./b", Import: []string{"./c"}}},
        {Key: "c", Package: workspace.Package{Name: "c", Version: "0.0.1", Root: "./c", Import: []string{"./a"}}},
    }
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    var buf bytes.Buffer
    _ = runLint(&buf, dir, true, false, false)

    dec := json.NewDecoder(&buf)
    var cycles [][]string
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        if m["code"] == "E_IMPORT_CYCLE" {
            data, _ := m["data"].(map[string]any)
            raw, _ := data["cycle"].([]any)
            var cyc []string
            for _, v := range raw { if s, ok := v.(string); ok { cyc = append(cyc, s) } }
            cycles = append(cycles, cyc)
        }
    }
    if len(cycles) != 1 { t.Fatalf("expected exactly 1 cycle, got %d: %+v", len(cycles), cycles) }
    cyc := cycles[0]
    if len(cyc) != 3 { t.Fatalf("expected 3-node cycle, got %d: %+v", len(cyc), cyc) }
    // Canonicalization: cycle should begin with lexicographically smallest node "./a"
    if cyc[0] != "./a" { t.Fatalf("expected canonicalized cycle starting with ./a, got %+v", cyc) }
    // Must contain ./b and ./c
    haveB, haveC := false, false
    for _, n := range cyc { if n == "./b" { haveB = true }; if n == "./c" { haveC = true } }
    if !haveB || !haveC { t.Fatalf("cycle missing nodes: %+v", cyc) }
}
