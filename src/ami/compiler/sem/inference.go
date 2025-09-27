package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeTypeInference performs M1 local type inference for identifiers and
// unary/binary expressions inside function bodies, emitting diagnostics with
// precise positions on mismatches and unknowns.
func AnalyzeTypeInference(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        // build local env
        env := map[string]string{}
        for _, p := range fn.Params {
            if p.Name != "" && p.Type != "" { env[p.Name] = p.Type }
        }
        // one pass to seed types from var decls with explicit types
        for _, st := range fn.Body.Stmts {
            if vd, ok := st.(*ast.VarDecl); ok {
                if vd.Name != "" {
                    if vd.Type != "" { env[vd.Name] = vd.Type } else if vd.Init != nil { env[vd.Name] = inferExprType(env, vd.Init) }
                }
            }
        }
        // second pass: assignments and mismatches
        for _, st := range fn.Body.Stmts {
            switch v := st.(type) {
            case *ast.AssignStmt:
                vt := inferExprType(env, v.Value)
                if old, ok := env[v.Name]; ok && old != "" && vt != "any" && old != vt {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "assignment type mismatch: expected " + old + ", got " + vt, Pos: &diag.Position{Line: v.NamePos.Line, Column: v.NamePos.Column, Offset: v.NamePos.Offset}})
                }
                if vt != "" && vt != "any" { env[v.Name] = vt }
            case *ast.ReturnStmt:
                // ensure returns match declared results when both sides are known (delegated to other analyzer);
                // here, we only try to detect ambiguous return literals.
                for _, e := range v.Results {
                    // bubble-up ambiguity
                    out = append(out, ambiguousInExpr(now, e)...)
                }
            }
        }
    }
    return out
}

func inferExprType(env map[string]string, e ast.Expr) string {
    switch v := e.(type) {
    case *ast.IdentExpr:
        if t := env[v.Name]; t != "" { return t }
        return "any"
    case *ast.NumberLit:
        return "int"
    case *ast.StringLit:
        return "string"
    case *ast.BinaryExpr:
        xt := inferExprType(env, v.X)
        yt := inferExprType(env, v.Y)
        switch v.Op {
        case token.Plus:
            if xt == "string" && yt == "string" { return "string" }
            if xt == "int" && yt == "int" { return "int" }
        case token.Minus, token.Star, token.Slash:
            if xt == "int" && yt == "int" { return "int" }
        default:
            if xt == yt && xt != "any" { return xt }
        }
        return "any"
    case *ast.SliceLit:
        if v.TypeName != "" { return "slice<" + v.TypeName + ">" }
        if len(v.Elems) == 0 { return "slice<any>" }
        et := inferExprType(env, v.Elems[0])
        if et == "" { et = "any" }
        return "slice<" + et + ">"
    case *ast.SetLit:
        if v.TypeName != "" { return "set<" + v.TypeName + ">" }
        if len(v.Elems) == 0 { return "set<any>" }
        et := inferExprType(env, v.Elems[0])
        if et == "" { et = "any" }
        return "set<" + et + ">"
    case *ast.MapLit:
        kt := v.KeyType
        vt := v.ValType
        if kt == "" { kt = "any" }
        if vt == "" { vt = "any" }
        return "map<" + kt + "," + vt + ">"
    default:
        return "any"
    }
}
