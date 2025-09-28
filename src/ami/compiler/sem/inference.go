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
        // collect local function result signatures for call propagation
        sigs := collectFunctionResults(f)
        // seed from var decls
        for _, st := range fn.Body.Stmts {
            if vd, ok := st.(*ast.VarDecl); ok {
                if vd.Name != "" {
                    if vd.Type != "" {
                        env[vd.Name] = vd.Type
                    } else if vd.Init != nil {
                        env[vd.Name] = inferLocalExprTypeWithSigs(env, sigs, vd.Init)
                    }
                }
            }
        }
        // conservative multi-pass propagation over assignments to reach a fixed point quickly
        for pass := 0; pass < 3; pass++ {
            changed := false
            for _, st := range fn.Body.Stmts {
                as, ok := st.(*ast.AssignStmt)
                if !ok { continue }
                vt := inferLocalExprTypeWithSigs(env, sigs, as.Value)
                if vt == "" || vt == "any" { continue }
                old := env[as.Name]
                if old == "" || old == "any" {
                    env[as.Name] = vt
                    changed = true
                    continue
                }
                // if both known but container generics unify, keep the old to avoid churn
                // otherwise leave as-is; mismatch will be reported below
            }
            if !changed { break }
        }
        // diagnostics: type mismatches and ambiguous returns
        for _, st := range fn.Body.Stmts {
            switch v := st.(type) {
            case *ast.AssignStmt:
                vt := inferLocalExprTypeWithSigs(env, sigs, v.Value)
                if old, ok := env[v.Name]; ok && old != "" {
                    if !typesCompatible(old, vt) {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "assignment type mismatch: expected " + old + ", got " + vt, Pos: &diag.Position{Line: v.NamePos.Line, Column: v.NamePos.Column, Offset: v.NamePos.Offset}})
                    }
                }
            case *ast.ReturnStmt:
                for _, e := range v.Results { out = append(out, ambiguousInExpr(now, e)...)}
            }
        }
    }
    return out
}

func inferLocalExprType(env map[string]string, e ast.Expr) string {
    switch v := e.(type) {
    case *ast.IdentExpr:
        if t := env[v.Name]; t != "" { return t }
        return "any"
    case *ast.NumberLit:
        return "int"
    case *ast.StringLit:
        return "string"
    case *ast.UnaryExpr:
        // logical not yields bool (i1)
        if v.Op == token.Bang { return "bool" }
        return "any"
    case *ast.BinaryExpr:
        xt := inferLocalExprType(env, v.X)
        yt := inferLocalExprType(env, v.Y)
        switch v.Op {
        case token.Plus:
            if xt == "string" && yt == "string" { return "string" }
            if xt == "int" && yt == "int" { return "int" }
        case token.Minus, token.Star, token.Slash:
            if xt == "int" && yt == "int" { return "int" }
        case token.Eq, token.Ne, token.Lt, token.Le, token.Gt, token.Ge:
            // comparisons yield bool regardless of operand type (when comparable)
            return "bool"
        case token.And, token.Or:
            // logical ops yield bool
            return "bool"
        default:
            if xt == yt && xt != "any" { return xt }
        }
        return "any"
    case *ast.SliceLit:
        if v.TypeName != "" { return "slice<" + v.TypeName + ">" }
        if len(v.Elems) == 0 { return "slice<any>" }
        et := inferLocalExprType(env, v.Elems[0])
        if et == "" { et = "any" }
        return "slice<" + et + ">"
    case *ast.SetLit:
        if v.TypeName != "" { return "set<" + v.TypeName + ">" }
        if len(v.Elems) == 0 { return "set<any>" }
        et := inferLocalExprType(env, v.Elems[0])
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

// inferExprType is an adapter used by other analyzers that need env-aware inference.
func inferExprType(env map[string]string, e ast.Expr) string { return inferLocalExprType(env, e) }

// inferLocalExprTypeWithSigs extends local inference to consult known function
// signatures for call expressions, returning the first result type when known.
func inferLocalExprTypeWithSigs(env map[string]string, sigs map[string][]string, e ast.Expr) string {
    switch v := e.(type) {
    case *ast.CallExpr:
        if rs, ok := sigs[v.Name]; ok && len(rs) > 0 && rs[0] != "" { return rs[0] }
        return "any"
    default:
        return inferLocalExprType(env, e)
    }
}

// collectFunctionResults builds a map of function name to declared result types.
func collectFunctionResults(f *ast.File) map[string][]string {
    out := map[string][]string{}
    if f == nil { return out }
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok {
            if len(fn.Results) == 0 { continue }
            rs := make([]string, len(fn.Results))
            for i, r := range fn.Results { rs[i] = r.Type }
            out[fn.Name] = rs
        }
    }
    return out
}
