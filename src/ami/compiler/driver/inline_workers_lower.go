package driver

import (
    "strings"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerInlineWorkers scans pipeline steps for inline func-literal workers and
// synthesizes named IR functions with deterministic names so that codegen can
// emit worker core wrappers. The generated body returns the input Event<T> and
// a nil error, which preserves types and enables execution via invoker path.
func lowerInlineWorkers(pkg, unit string, f *ast.File) []ir.Function {
    var out []ir.Function
    if f == nil { return out }
    // For each pipeline independently, count inline occurrences deterministically
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        inlineIdx := 0
        for _, s := range pd.Stmts {
            st, ok := s.(*ast.StepStmt)
            if !ok { continue }
            lname := strings.ToLower(st.Name)
            if lname != "transform" && lname != "ingress" && lname != "egress" { continue }
            // find worker arg (named or first positional)
            var warg string
            has := false
            for _, a := range st.Args {
                if eq := strings.IndexByte(a.Text, '='); eq > 0 {
                    key := strings.TrimSpace(a.Text[:eq])
                    if strings.EqualFold(key, "worker") {
                        warg = strings.TrimSpace(a.Text[eq+1:])
                        has = true
                        break
                    }
                } else if !has {
                    warg = a.Text
                    has = true
                }
            }
            if !has { continue }
            if !strings.HasPrefix(strings.TrimSpace(warg), "func") { continue }
            // Parse signature minimally from the literal text
            paramTyp, results, ok := semInlineSig(strings.TrimSpace(warg))
            if !ok { continue }
            tParam := extractEventArg(paramTyp)
            if tParam == "" { tParam = "any" }
            tResult := tParam
            if len(results) > 0 { if tr := extractEventArg(strings.TrimSpace(results[0])); tr != "" { tResult = tr } }
            // Deterministic name consistent with pipelines debug normalization
            inlineIdx++
            name := "InlineWorker_" + unit + "_" + pd.Name + "_" + itoa(inlineIdx)
            // Build IR function: func(ev Event<T>) (Event<U>, error) { return ev, nil }
            evTy := "Event<" + tParam + ">"
            r0Ty := "Event<" + tResult + ">"
            fn := ir.Function{Name: name}
            fn.Params = []ir.Value{{ID: "ev", Type: evTy}}
            fn.Results = []ir.Value{{ID: "r0", Type: r0Ty}, {ID: "r1", Type: "error"}}
            // err0 := null pointer (via fallback field expr)
            errID := "err0"
            fn.Blocks = []ir.Block{{Name: "entry", Instr: []ir.Instruction{
                ir.Expr{Op: "field.null", Result: &ir.Value{ID: errID, Type: "error"}},
                ir.Return{Values: []ir.Value{{ID: "ev", Type: evTy}, {ID: errID, Type: "error"}}},
            }}}
            out = append(out, fn)
        }
    }
    return out
}
