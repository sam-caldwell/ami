package driver

import (
    "encoding/json"
    "os"
    "testing"
)

func TestWriteEventMetaDebug_WritesScaffold(t *testing.T) {
    path, err := writeEventMetaDebug("main", "u1")
    if err != nil { t.Fatalf("write: %v", err) }
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read: %v", err) }
    var em struct{ Schema, Package, Unit string; ImmutablePayload bool }
    if err := json.Unmarshal(b, &em); err != nil { t.Fatalf("json: %v", err) }
    if em.Schema != "eventmeta.v1" || em.Package != "main" || em.Unit != "u1" || !em.ImmutablePayload {
        t.Fatalf("unexpected eventmeta: %+v", em)
    }
}

