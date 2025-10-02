package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
)

// lowerSelectorField attempts to resolve a selector expression as a field projection
// using the receiver's declared type from the local environment. If it can
// determine the resulting field type, it emits a synthetic "field.<name>" expr
// that defines a new SSA value of the appropriate type.
func lowerSelectorField(st *lowerState, s *ast.SelectorExpr) (ir.Expr, bool) {
    if st == nil || s == nil { return ir.Expr{}, false }
    // Flatten selector chain to base ident and path: base.a.b.c
    baseIdent, path := flattenSelector(s)
    if baseIdent == "" || path == "" { return ir.Expr{}, false }
    // Resolve base type from local var types
    btype := st.varTypes[baseIdent]
    if btype == "" { return ir.Expr{}, false }
    // Parse and resolve field path via types package
    root, err := types.Parse(btype)
    if err != nil { return ir.Expr{}, false }
    ft, ok := types.ResolveField(root, path)
    if !ok || ft == nil { return ir.Expr{}, false }
    fts := ft.String()
    // Determine result IR type
    rtype := fts
    // Normalize Time to int64 handle for runtime ABI friendliness
    if fts == "Time" { rtype = "int64" }
    id := st.newTemp()
    res := &ir.Value{ID: id, Type: rtype}
    // Provide the base as an argument for potential future codegen; keep original
    // base type text so codegen can compute field offsets/layout.
    bargType := st.varTypes[baseIdent]
    arg := ir.Value{ID: baseIdent, Type: bargType}
    // Encode field name into the op for debug purposes: field.<path>
    return ir.Expr{Op: "field." + path, Args: []ir.Value{arg}, Result: res}, true
}

