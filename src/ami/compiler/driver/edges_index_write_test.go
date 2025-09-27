package driver

import (
    "bytes"
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
    if !bytes.Contains(b, []byte(`"schema":"edges.v1"`)) { t.Fatalf("missing schema: %s", string(b)) }
    // Ensure a->b appears before b->c in file contents to validate sorting
    ai := bytes.Index(b, []byte(`"from":"a","to":"b"`))
    bi := bytes.Index(b, []byte(`"from":"b","to":"c"`))
    if ai < 0 || bi < 0 || !(ai < bi) { t.Fatalf("order not sorted: %s", string(b)) }
}

