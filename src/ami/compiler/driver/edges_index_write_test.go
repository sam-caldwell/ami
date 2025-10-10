package driver

import (
    "encoding/json"
    "os"
    "testing"
)

func TestWriteEdgesIndex_SortsAndWrites(t *testing.T) {
    edges := []edgeEntry{
        {Unit: "b", From: "b", To: "c"},
        {Unit: "a", From: "a", To: "b"},
    }
    path, err := writeEdgesIndex("main", edges, nil)
    if err != nil { t.Fatalf("write: %v", err) }
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read: %v", err) }
    var idx edgesIndex
    if err := json.Unmarshal(b, &idx); err != nil { t.Fatalf("json: %v", err) }
    if idx.Schema != "edges.v1" || idx.Package != "main" { t.Fatalf("unexpected header: %+v", idx) }
    if len(idx.Edges) < 2 { t.Fatalf("expected at least 2 edges: %+v", idx.Edges) }
    if !(idx.Edges[0].From == "a" && idx.Edges[0].To == "b") { t.Fatalf("first edge not a->b: %+v", idx.Edges[0]) }
    if !(idx.Edges[1].From == "b" && idx.Edges[1].To == "c") { t.Fatalf("second edge not b->c: %+v", idx.Edges[1]) }
}

func TestWriteEdgesIndex_IncludesCapacityBackpressure(t *testing.T) {
    edges := []edgeEntry{{Unit: "u", Pipeline: "P", From: "ingress", To: "Transform", MinCapacity: 1, MaxCapacity: 5, Backpressure: "dropOldest"}}
    path, err := writeEdgesIndex("main", edges, nil)
    if err != nil { t.Fatalf("write: %v", err) }
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read: %v", err) }
    var idx map[string]any
    if err := json.Unmarshal(b, &idx); err != nil { t.Fatalf("json: %v", err) }
    es := idx["edges"].([]any)
    if len(es) == 0 { t.Fatalf("edges missing") }
    e := es[0].(map[string]any)
    if e["minCapacity"].(float64) != 1 || e["maxCapacity"].(float64) != 5 || e["backpressure"].(string) != "dropOldest" {
        t.Fatalf("unexpected fields: %+v", e)
    }
}
