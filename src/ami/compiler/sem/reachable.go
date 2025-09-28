package sem

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// ReachableFunctions computes a set of function names reachable from roots within a file.
// Roots: functions named "main"; any function referenced by a call; defer; or in return.
func ReachableFunctions(f *ast.File) map[string]bool {
    if f == nil { return map[string]bool{} }
    defs := map[string]bool{}
    for _, d := range f.Decls { if fn, ok := d.(*ast.FuncDecl); ok { defs[fn.Name] = true } }
    refs := map[string]bool{}
    refs["main"] = true
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        for _, st := range fn.Body.Stmts {
            switch v := st.(type) {
            case *ast.ExprStmt:
                if ce, ok := v.X.(*ast.CallExpr); ok { refs[ce.Name] = true }
            case *ast.DeferStmt:
                if v.Call != nil { refs[v.Call.Name] = true }
            case *ast.ReturnStmt:
                for _, e := range v.Results { if ce, ok := e.(*ast.CallExpr); ok { refs[ce.Name] = true } }
            }
        }
    }
    // intersect refs with defs, and include main if present
    reach := map[string]bool{}
    for name := range defs { if refs[name] { reach[name] = true } }
    return reach
}

