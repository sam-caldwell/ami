package ir

import (
    "encoding/json"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestIR_FunctionTypeParams_Constraints_RoundTrip(t *testing.T) {
    src := `package p
func g<T any, U any>(a T, b U) { }`
    p := parser.New(src)
    f := p.ParseFile()
    m := FromASTFile("p", "", "unit.ami", f)
    ir := m.ToSchema()
    // Marshal/unmarshal to simulate artifacts
    b, err := json.Marshal(ir)
    if err != nil { t.Fatalf("marshal: %v", err) }
    // decode minimal subset (anonymous struct) of schemas.IRV1.Functions
    var v struct{ Functions []struct{ Name string; TypeParams []struct{ Name, Constraint string } } }
    if err := json.Unmarshal(b, &v); err != nil { t.Fatalf("unmarshal: %v", err) }
    // Find g
    var tp []struct{ Name, Constraint string }
    for _, fn := range v.Functions { if fn.Name == "g" { tp = fn.TypeParams; break } }
    if len(tp) != 2 || tp[0].Name != "T" || tp[0].Constraint != "any" || tp[1].Name != "U" || tp[1].Constraint != "any" {
        t.Fatalf("unexpected type params: %+v", tp)
    }
}
