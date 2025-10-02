package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeCallsWithSigs is like AnalyzeCalls but uses the provided package-wide signature maps,
// and optionally parameter type positions from the driver. When positions are not provided,
// it falls back to local function declarations in the same file.
func AnalyzeCallsWithSigs(f *ast.File, params map[string][]string, results map[string][]string, paramPos map[string][]diag.Position, paramNames map[string][]string) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // Gather local param type positions for functions present in this file (best-effort)
    localParamPos := map[string][]diag.Position{}
    localParamNames := map[string][]string{}
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok {
            var pos []diag.Position
            var names []string
            for _, p := range fn.Params {
                tp := diag.Position{Line: p.TypePos.Line, Column: p.TypePos.Column, Offset: p.TypePos.Offset}
                if p.TypePos.Line == 0 { tp = diag.Position{Line: p.Pos.Line, Column: p.Pos.Column, Offset: p.Pos.Offset} }
                pos = append(pos, tp)
                names = append(names, p.Name)
            }
            localParamPos[fn.Name] = pos
            localParamNames[fn.Name] = names
        }
    }
    // analyze each function with local var types
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        // build local env (params, var inits/types, assignments)
        vars := buildLocalEnv(fn)
        for _, st := range fn.Body.Stmts {
            switch v := st.(type) {
            case *ast.ExprStmt:
                if ce, ok := v.X.(*ast.CallExpr); ok {
                    // prefer driver-provided positions when present
                    effective := localParamPos
                    if paramPos != nil { effective = paramPos }
                    effNames := localParamNames
                    if paramNames != nil { effNames = paramNames }
                    out = append(out, checkCallWithSigsWithResults(ce, params, results, now, vars, effective, effNames)...)
                }
            case *ast.DeferStmt:
                if v.Call != nil {
                    effective := localParamPos
                    if paramPos != nil { effective = paramPos }
                    effNames := localParamNames
                    if paramNames != nil { effNames = paramNames }
                    out = append(out, checkCallWithSigsWithResults(v.Call, params, results, now, vars, effective, effNames)...)
                }
            case *ast.ReturnStmt:
                for _, e := range v.Results {
                    if ce, ok := e.(*ast.CallExpr); ok {
                        effective := localParamPos
                        if paramPos != nil { effective = paramPos }
                        effNames := localParamNames
                        if paramNames != nil { effNames = paramNames }
                        out = append(out, checkCallWithSigsWithResults(ce, params, results, now, vars, effective, effNames)...)
                    }
                }
            }
        }
    }
    return out
}

