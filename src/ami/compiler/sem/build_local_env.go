package sem

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// buildLocalEnv collects local variable types from parameters, var declarations,
// and simple assignments where the right-hand side has a deducible type.
func buildLocalEnv(fn *ast.FuncDecl) map[string]string {
    env := map[string]string{}
    for _, p := range fn.Params { if p.Name != "" && p.Type != "" { env[p.Name] = p.Type } }
    if fn.Body == nil { return env }
    for _, st := range fn.Body.Stmts {
        switch v := st.(type) {
        case *ast.VarDecl:
            if v.Name != "" {
                if v.Type != "" { env[v.Name] = v.Type } else if v.Init != nil {
                    if t := deduceType(v.Init); t != "any" && t != "" { env[v.Name] = t }
                }
            }
        case *ast.AssignStmt:
            if v.Name != "" && v.Value != nil {
                if t := deduceType(v.Value); t != "any" && t != "" { env[v.Name] = t }
            }
        }
    }
    return env
}

