package ir

import (
    "testing"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// Verify IR function param annotations include ownership and domain.
func TestIR_ToSchema_ParamAnnotations_OwnershipAndDomain(t *testing.T) {
    // Construct an AST with a single function and a variety of param types.
    fd := astpkg.FuncDecl{
        Name: "f",
        Params: []astpkg.Param{
            {Name: "ctx", Type: astpkg.TypeRef{Name: "Context"}},
            {Name: "ev", Type: astpkg.TypeRef{Name: "Event", Args: []astpkg.TypeRef{{Name: "string"}}}},
            {Name: "r", Type: astpkg.TypeRef{Name: "Owned", Args: []astpkg.TypeRef{{Name: "int"}}}},
            {Name: "st", Type: astpkg.TypeRef{Name: "State", Ptr: true}},
            {Name: "st2", Type: astpkg.TypeRef{Name: "State"}},
            {Name: "x", Type: astpkg.TypeRef{Name: "int"}},
        },
    }
    f := &astpkg.File{Package: "p", Decls: []astpkg.Node{fd}}
    m := FromASTFile("p", "unit.ami", f)
    ir := m.ToSchema()
    if len(ir.Functions) != 1 { t.Fatalf("functions=%d", len(ir.Functions)) }
    params := ir.Functions[0].Params
    if len(params) != 6 { t.Fatalf("params=%d", len(params)) }
    // Build a quick lookup by name for stable assertions.
    byName := map[string]schParam{}
    for _, p := range params { byName[p.Name] = schParam{Domain:p.Domain, Ownership:p.Ownership} }
    // Event param → domain=event, ownership=borrowed
    if v := byName["ev"]; v.Domain != "event" || v.Ownership != "borrowed" { t.Fatalf("event annotations: %+v", v) }
    // Owned<T> param → ownership=owned, domain=ephemeral by default
    if v := byName["r"]; v.Ownership != "owned" || v.Domain != "ephemeral" { t.Fatalf("owned annotations: %+v", v) }
    // State param (ptr or not) → domain=state
    if v := byName["st"]; v.Domain != "state" { t.Fatalf("state(ptr) annotation: %+v", v) }
    if v := byName["st2"]; v.Domain != "state" { t.Fatalf("state(val) annotation: %+v", v) }
    // Ephemeral default for regular params
    if v := byName["x"]; v.Domain != "ephemeral" { t.Fatalf("ephemeral annotation: %+v", v) }
}

// minimal struct to avoid importing schemas here
type schParam struct{ Domain, Ownership string }
