package ir

import (
    "encoding/json"
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
    return obj
}
