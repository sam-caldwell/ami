package ir

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
)

// EncodeModule produces deterministic JSON for debug.
func EncodeModule(m Module) ([]byte, error) {
    // Build a JSON-friendly shape explicitly to control ordering.
    jm := map[string]any{
        "schema":   "ir.v1",
        "package":  m.Package,
        "functions": []any{},
    }
    fns := make([]any, 0, len(m.Functions))
    for _, f := range m.Functions {
        jf := map[string]any{
            "name":    f.Name,
            "params":  valuesToJSON(f.Params),
            "results": valuesToJSON(f.Results),
            "blocks":  []any{},
        }
        bl := make([]any, 0, len(f.Blocks))
        for _, b := range f.Blocks {
            jb := map[string]any{
                "name":   b.Name,
                "instrs": instrsToJSON(b.Instr),
            }
            bl = append(bl, jb)
        }
        jf["blocks"] = bl
        fns = append(fns, jf)
    }
    jm["functions"] = fns
    return json.MarshalIndent(jm, "", "  ")
}

// WriteDebug writes module JSON under build/debug/ir/<pkg>.json
func WriteDebug(m Module) error {
    path := filepath.Join("build", "debug", "ir")
    if err := os.MkdirAll(path, 0o755); err != nil { return err }
    b, err := EncodeModule(m)
    if err != nil { return err }
    fname := filepath.Join(path, m.Package+".json")
    return os.WriteFile(fname, b, 0o644)
}

func valuesToJSON(vs []Value) []any {
    out := make([]any, 0, len(vs))
    for _, v := range vs {
        out = append(out, map[string]any{"id": v.ID, "type": v.Type})
    }
    return out
}

func instrsToJSON(ins []Instruction) []any {
    out := make([]any, 0, len(ins))
    for _, in := range ins {
        switch v := in.(type) {
        case Var:
            obj := map[string]any{"op": OpVar.String(), "name": v.Name, "type": v.Type, "result": map[string]any{"id": v.Result.ID, "type": v.Result.Type}}
            if v.Init != nil { obj["init"] = map[string]any{"id": v.Init.ID, "type": v.Init.Type} }
            out = append(out, obj)
        case Assign:
            out = append(out, map[string]any{"op": OpAssign.String(), "dest": v.DestID, "src": map[string]any{"id": v.Src.ID, "type": v.Src.Type}})
        case Return:
            vals := valuesToJSON(v.Values)
            out = append(out, map[string]any{"op": OpReturn.String(), "values": vals})
        case Defer:
            out = append(out, map[string]any{"op": OpDefer.String(), "expr": exprToJSON(v.Expr)})
        case Expr:
            out = append(out, map[string]any{"op": OpExpr.String(), "expr": exprToJSON(v)})
        default:
            // ignore unknown
        }
    }
    return out
}

func exprToJSON(e Expr) any {
    obj := map[string]any{"op": e.Op}
    if e.Callee != "" { obj["callee"] = e.Callee }
    if len(e.Args) > 0 { obj["args"] = valuesToJSON(e.Args) }
    if e.Result != nil { obj["result"] = map[string]any{"id": e.Result.ID, "type": e.Result.Type} }
    // Enrich call expressions with simple type signature hints for downstream phases.
    if e.Op == "call" {
        if len(e.Args) > 0 {
            ats := make([]any, 0, len(e.Args))
            for _, a := range e.Args { ats = append(ats, a.Type) }
            obj["argTypes"] = ats
        } else {
            obj["argTypes"] = []any{}
        }
        if e.Result != nil {
            obj["retTypes"] = []any{e.Result.Type}
        } else {
            obj["retTypes"] = []any{}
        }
        // Always include a signature block for calls in debug JSON
        sig := map[string]any{"params": []any{}, "results": []any{}}
        var ps []any
        // Case 1: have names matching types → emit name/type objects
        if len(e.ParamNames) > 0 && len(e.ParamNames) == len(e.ParamTypes) {
            ps = make([]any, 0, len(e.ParamTypes))
            for i := range e.ParamTypes {
                ps = append(ps, map[string]any{"name": e.ParamNames[i], "type": e.ParamTypes[i]})
            }
        } else {
            // Case 2: synthesize placeholders to always provide name/type objects
            // Prefer function param types; otherwise fall back to arg types
            n := len(e.ParamTypes)
            if n == 0 { n = len(e.Args) }
            ps = make([]any, 0, n)
            for i := 0; i < n; i++ {
                name := fmt.Sprintf("p%d", i)
                typ := "any"
                if i < len(e.ParamTypes) && e.ParamTypes[i] != "" { typ = e.ParamTypes[i] }
                if typ == "any" && i < len(e.Args) && e.Args[i].Type != "" { typ = e.Args[i].Type }
                ps = append(ps, map[string]any{"name": name, "type": typ})
            }
        }
        rs := make([]any, 0, len(e.ResultTypes))
        if len(e.ResultTypes) > 0 {
            for _, r := range e.ResultTypes { rs = append(rs, r) }
        } else if e.Result != nil {
            rs = append(rs, e.Result.Type)
        }
        sig["params"] = ps
        sig["results"] = rs
        obj["sig"] = sig
    }
    return obj
}
