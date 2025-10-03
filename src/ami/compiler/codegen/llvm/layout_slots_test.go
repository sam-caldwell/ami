package llvm

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
)

func Test_slotsOf_Primitive_Optional_Union_Struct(t *testing.T) {
    // Primitive → 1
    if s := slotsOf(types.Primitive{K: types.Int}); s != 1 { t.Fatalf("prim: %d", s) }
    // Optional(inner) → slots(inner)
    if s := slotsOf(types.Optional{Inner: types.Primitive{K: types.Bool}}); s != 1 { t.Fatalf("opt: %d", s) }
    // Union(alts) → max(slots(alts))
    u := types.Union{Alts: []types.Type{types.Primitive{K: types.Int}, types.Primitive{K: types.Bool}}}
    if s := slotsOf(u); s != 1 { t.Fatalf("union: %d", s) }
    // Struct fields sum, stable by name
    st := types.Struct{Fields: map[string]types.Type{
        "b": types.Primitive{K: types.Int},
        "a": types.Primitive{K: types.Bool},
    }}
    if s := slotsOf(st); s != 2 { t.Fatalf("struct: %d", s) }
}

