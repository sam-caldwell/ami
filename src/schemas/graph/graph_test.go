package graph

import (
    "bytes"
    "encoding/json"
    "testing"
)

func TestGraph_MarshalJSON_ContainsSchemaAndSorted(t *testing.T) {
    g := Graph{
        Package: "main",
        Unit:    "pipeline.ami",
        Name:    "Demo",
        Nodes: []Node{
            {ID: "b", Kind: "worker", Label: "B"},
            {ID: "a", Kind: "ingress", Label: "A"},
        },
        Edges: []Edge{
            {From: "b", To: "c"},
            {From: "a", To: "b"},
        },
    }
    b, err := json.Marshal(g)
    if err != nil { t.Fatalf("marshal: %v", err) }
    if !bytes.Contains(b, []byte(`"schema":"graph.v1"`)) {
        t.Fatalf("missing schema: %s", string(b))
    }
    // ensure nodes sorted by id: a then b
    ai := bytes.Index(b, []byte(`"id":"a"`))
    bi := bytes.Index(b, []byte(`"id":"b"`))
    if ai < 0 || bi < 0 || !(ai < bi) { t.Fatalf("nodes not sorted: %s", string(b)) }
    // ensure edges sorted by from/to: a->b appears before b->c
    aEdge := bytes.Index(b, []byte(`"from":"a","to":"b"`))
    bEdge := bytes.Index(b, []byte(`"from":"b","to":"c"`))
    if aEdge < 0 || bEdge < 0 || !(aEdge < bEdge) { t.Fatalf("edges not sorted: %s", string(b)) }
}

