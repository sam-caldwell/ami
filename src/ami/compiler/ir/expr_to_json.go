package ir

import "fmt"

func exprToJSON(e Expr) any {
    obj := map[string]any{"op": e.Op}
    if e.Callee != "" { obj["callee"] = e.Callee }
    if len(e.Args) > 0 { obj["args"] = valuesToJSON(e.Args) }
    if e.Result != nil { obj["result"] = map[string]any{"id": e.Result.ID, "type": e.Result.Type} }
    if len(e.Results) > 0 { obj["results"] = valuesToJSON(e.Results) }
    if e.Op == "call" {
        if len(e.Args) > 0 {
            ats := make([]any, 0, len(e.Args))
            for _, a := range e.Args { ats = append(ats, a.Type) }
            obj["argTypes"] = ats
        } else { obj["argTypes"] = []any{} }
        if len(e.Results) > 0 {
            rt := make([]any, 0, len(e.Results))
            for _, r := range e.Results { rt = append(rt, r.Type) }
            obj["retTypes"] = rt
        } else if e.Result != nil { obj["retTypes"] = []any{e.Result.Type} } else { obj["retTypes"] = []any{} }
        sig := map[string]any{"params": []any{}, "results": []any{}}
        var ps []any
        if len(e.ParamNames) > 0 && len(e.ParamNames) == len(e.ParamTypes) {
            ps = make([]any, 0, len(e.ParamTypes))
            for i := range e.ParamTypes { ps = append(ps, map[string]any{"name": e.ParamNames[i], "type": e.ParamTypes[i]}) }
        } else {
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
        if len(e.ResultTypes) > 0 { for _, r := range e.ResultTypes { rs = append(rs, r) } } else if len(e.Results) > 0 { for _, r := range e.Results { rs = append(rs, r.Type) } } else if e.Result != nil { rs = append(rs, e.Result.Type) }
        sig["params"] = ps
        sig["results"] = rs
        obj["sig"] = sig
    }
    return obj
}

