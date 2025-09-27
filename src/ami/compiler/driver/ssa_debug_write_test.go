package driver

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestWriteSSADebug_VersionsDefs(t *testing.T) {
    // Build a tiny IR module with two defs of the same variable
    f := ir.Function{
        Name:   "F",
        Blocks: []ir.Block{{
            Name: "entry",
            Instr: []ir.Instruction{
                ir.Var{Name: "x", Type: "int", Result: ir.Value{ID: "x0", Type: "int"}},
                ir.Assign{DestID: "x", Src: ir.Value{ID: "c1", Type: "int"}},
                ir.Return{Values: []ir.Value{{ID: "x", Type: "int"}}},
            },
        }},
    }
    m := ir.Module{Package: "app", Functions: []ir.Function{f}}
    p, err := writeSSADebug("app", "u", m)
    if err != nil { t.Fatalf("writeSSA: %v", err) }
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if obj["schema"] != "ssa.v1" { t.Fatalf("schema: %v", obj["schema"]) }
    funs := obj["functions"].([]any)
    if len(funs) != 1 { t.Fatalf("functions len: %d", len(funs)) }
    defs := funs[0].(map[string]any)["defs"].([]any)
    if len(defs) < 2 { t.Fatalf("expected >=2 defs, got %d", len(defs)) }
    d0 := defs[0].(map[string]any)
    d1 := defs[1].(map[string]any)
    if d0["ssaName"].(string) != "x#0" || d1["ssaName"].(string) != "x#1" {
        t.Fatalf("unexpected ssa names: %s, %s", d0["ssaName"], d1["ssaName"])
    }
}
