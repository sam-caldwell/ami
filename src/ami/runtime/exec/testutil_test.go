package exec

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// WriteEdges writes build/debug/asm/<pkg>/edges.json with the provided entries.
func WriteEdges(t *testing.T, pkg, pipeline string, edges []edgeEntry) {
    t.Helper()
    dir := filepath.Join("build", "debug", "asm", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    idx := edgesIndex{Schema: "asm.v1", Package: pkg, Edges: edges}
    b, _ := json.Marshal(idx)
    if err := os.WriteFile(filepath.Join(dir, "edges.json"), b, 0o644); err != nil { t.Fatalf("write edges: %v", err) }
}

// MakeModuleWithEdges constructs a minimal Module and writes edges for tests.
func MakeModuleWithEdges(t *testing.T, pkg, pipeline string, edges []edgeEntry) ir.Module {
    t.Helper()
    WriteEdges(t, pkg, pipeline, edges)
    return ir.Module{Package: pkg, Pipelines: []ir.Pipeline{{Name: pipeline}}}
}

