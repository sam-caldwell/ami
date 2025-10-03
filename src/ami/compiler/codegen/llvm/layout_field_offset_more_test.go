package llvm

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
)

func Test_fieldOffsetSlots_NestedAndOptional(t *testing.T) {
    inner := types.Struct{Fields: map[string]types.Type{
        "x": types.Primitive{K: types.Int},
        "y": types.Primitive{K: types.Bool},
    }}
    root := types.Struct{Fields: map[string]types.Type{
        "a": inner,
        "b": types.Optional{Inner: types.Primitive{K: types.Int}},
    }}
    // offset of a.y: within inner struct, fields sorted => x(1 slot) before y, so off=1
    off, leaf, ok := fieldOffsetSlots(root, "a.y")
    if !ok || off != 1 { t.Fatalf("a.y off=%d ok=%v leaf=%v", off, ok, leaf) }
    // offset of b (optional) should be after a’s slots (2)
    off2, leaf2, ok2 := fieldOffsetSlots(root, "b")
    if !ok2 || off2 != 2 { t.Fatalf("b off=%d ok=%v leaf=%v", off2, ok2, leaf2) }
    // non-struct root
    if off, _, ok := fieldOffsetSlots(types.Primitive{K: types.Int}, "x"); ok || off != 0 {
        t.Fatalf("expected failure on non-struct root")
    }
    // unknown field
    if off, _, ok := fieldOffsetSlots(root, "a.z"); ok || off != 0 {
        t.Fatalf("expected failure on missing field")
    }
}
