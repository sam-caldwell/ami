package driver

import (
    "encoding/json"
    "os"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestWriteObjectStub_WritesObjV1(t *testing.T) {
    m := ir.Module{Functions: []ir.Function{{Name: "F"}}}
    p, err := writeObjectStub("main", "u1", m)
    if err != nil { t.Fatalf("write: %v", err) }
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read: %v", err) }
    var obj struct{ Schema string; Functions []string }
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if obj.Schema != "obj.v1" || len(obj.Functions) != 1 || obj.Functions[0] != "F" {
        t.Fatalf("unexpected: %+v", obj)
    }
}

