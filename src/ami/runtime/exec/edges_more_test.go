package exec

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestBuildLinearPathFromEdges_BasicPath(t *testing.T) {
    dir := t.TempDir()
    asm := filepath.Join(dir, "build", "debug", "asm", "app")
    if err := os.MkdirAll(asm, 0o755); err != nil { t.Fatal(err) }
    idx := edgesIndex{Schema: "asm.v1", Package: "app", Edges: []edgeEntry{
        {Pipeline: "P", From: "ingress", To: "Collect"},
        {Pipeline: "P", From: "Collect", To: "egress"},
    }}
    b, _ := json.Marshal(idx)
    if err := os.WriteFile(filepath.Join(asm, "edges.json"), b, 0o644); err != nil { t.Fatal(err) }
    nodes, err := BuildLinearPathFromEdges(dir, "app", "P")
    if err != nil { t.Fatalf("build path: %v", err) }
    if len(nodes) < 2 || nodes[0] != "ingress" || nodes[len(nodes)-1] != "egress" { t.Fatalf("unexpected nodes: %v", nodes) }
}

func TestBuildLinearPathFromEdges_MissingFileReturnsError(t *testing.T) {
    if _, err := BuildLinearPathFromEdges(t.TempDir(), "app", "P"); err == nil {
        t.Fatalf("expected error for missing edges.json")
    }
}

func TestBuildLinearPathFromEdges_DetectsLoop(t *testing.T) {
    dir := t.TempDir()
    asm := filepath.Join(dir, "build", "debug", "asm", "app")
    if err := os.MkdirAll(asm, 0o755); err != nil { t.Fatal(err) }
    idx := edgesIndex{Schema: "asm.v1", Package: "app", Edges: []edgeEntry{
        {Pipeline: "P", From: "ingress", To: "A"},
        {Pipeline: "P", From: "A", To: "A"},
    }}
    b, _ := json.Marshal(idx)
    if err := os.WriteFile(filepath.Join(asm, "edges.json"), b, 0o644); err != nil { t.Fatal(err) }
    nodes, err := BuildLinearPathFromEdges(dir, "app", "P")
    if err != nil { t.Fatalf("build path: %v", err) }
    if len(nodes) != 2 || nodes[1] != "A" { t.Fatalf("unexpected loop path: %v", nodes) }
}
