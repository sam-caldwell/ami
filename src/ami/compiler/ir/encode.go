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
    if m.Concurrency > 0 { jm["concurrency"] = m.Concurrency }
    if m.Backpressure != "" { jm["backpressurePolicy"] = m.Backpressure }
    if m.TelemetryEnabled { jm["telemetryEnabled"] = true }
    if m.Schedule != "" { jm["schedule"] = m.Schedule }
    if len(m.Capabilities) > 0 { jm["capabilities"] = m.Capabilities }
    if m.TrustLevel != "" { jm["trustLevel"] = m.TrustLevel }
    if m.ExecContext != nil { jm["execContext"] = m.ExecContext }
    if m.EventMeta != nil {
        jm["eventmeta"] = map[string]any{
            "schema": m.EventMeta.Schema,
            "fields": m.EventMeta.Fields,
        }
    }
    // directives (pragma-derived)
    if len(m.Directives) > 0 {
        ds := make([]any, 0, len(m.Directives))
        for _, d := range m.Directives {
            obj := map[string]any{"domain": d.Domain}
            if d.Key != "" { obj["key"] = d.Key }
            if d.Value != "" { obj["value"] = d.Value }
            if len(d.Args) > 0 { obj["args"] = d.Args }
            if len(d.Params) > 0 { obj["params"] = d.Params }
            ds = append(ds, obj)
        }
        jm["directives"] = ds
    }
    // pipelines (Collect/merge)
    if len(m.Pipelines) > 0 {
        ps := make([]any, 0, len(m.Pipelines))
        for _, p := range m.Pipelines {
            pj := map[string]any{"name": p.Name}
            if len(p.Collect) > 0 {
                cols := make([]any, 0, len(p.Collect))
                for _, c := range p.Collect {
                    cj := map[string]any{"step": c.Step}
                    if c.Merge != nil {
                        mj := map[string]any{}
                        if c.Merge.Stable { mj["stable"] = true }
                        if len(c.Merge.Sort) > 0 {
                            sk := make([]any, 0, len(c.Merge.Sort))
                            for _, s := range c.Merge.Sort { sk = append(sk, map[string]any{"field": s.Field, "order": s.Order}) }
                            mj["sort"] = sk
                        }
                        if c.Merge.Key != "" { mj["key"] = c.Merge.Key }
                        if c.Merge.PartitionBy != "" { mj["partitionBy"] = c.Merge.PartitionBy }
                        if c.Merge.Buffer.Capacity > 0 || c.Merge.Buffer.Policy != "" {
                            mj["buffer"] = map[string]any{"capacity": c.Merge.Buffer.Capacity, "policy": c.Merge.Buffer.Policy}
                        }
                        if c.Merge.Window > 0 { mj["window"] = c.Merge.Window }
                        if c.Merge.TimeoutMs > 0 { mj["timeoutMs"] = c.Merge.TimeoutMs }
                        if c.Merge.DedupField != "" { mj["dedupField"] = c.Merge.DedupField }
                        if c.Merge.Watermark != nil {
                            mj["watermark"] = map[string]any{"field": c.Merge.Watermark.Field, "latenessMs": c.Merge.Watermark.LatenessMs}
                        }
                        if c.Merge.LatePolicy != "" { mj["latePolicy"] = c.Merge.LatePolicy }
                        cj["merge"] = mj
                    }
                    cols = append(cols, cj)
                }
                pj["collect"] = cols
            }
            ps = append(ps, pj)
        }
        jm["pipelines"] = ps
    }

    fns := make([]any, 0, len(m.Functions))
    for _, f := range m.Functions {
        jf := map[string]any{
            "name":    f.Name,
            "params":  valuesToJSON(f.Params),
            "results": valuesToJSON(f.Results),
            "blocks":  []any{},
        }
        if len(f.Decorators) > 0 {
            decs := make([]any, 0, len(f.Decorators))
            for _, d := range f.Decorators {
                obj := map[string]any{"name": d.Name}
                if len(d.Args) > 0 { obj["args"] = d.Args }
                decs = append(decs, obj)
            }
            jf["decorators"] = decs
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
        case Phi:
            // Represent PHI with result and incomings
            inc := make([]any, 0, len(v.Incomings))
            for _, in := range v.Incomings { inc = append(inc, map[string]any{"value": map[string]any{"id": in.Value.ID, "type": in.Value.Type}, "label": in.Label}) }
            out = append(out, map[string]any{"op": OpPhi.String(), "result": map[string]any{"id": v.Result.ID, "type": v.Result.Type}, "incomings": inc})
        case CondBr:
            out = append(out, map[string]any{"op": OpCondBr.String(), "cond": map[string]any{"id": v.Cond.ID, "type": v.Cond.Type}, "true": v.TrueLabel, "false": v.FalseLabel})
        case Loop:
            out = append(out, map[string]any{"op": OpLoop.String(), "name": v.Name})
        case Goto:
            out = append(out, map[string]any{"op": OpGoto.String(), "label": v.Label})
        case SetPC:
            out = append(out, map[string]any{"op": OpSetPC.String(), "pc": v.PC})
        case Dispatch:
            out = append(out, map[string]any{"op": OpDispatch.String(), "label": v.Label})
        case PushFrame:
            out = append(out, map[string]any{"op": OpPushFrame.String(), "fn": v.Fn})
        case PopFrame:
            out = append(out, map[string]any{"op": OpPopFrame.String()})
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
    if len(e.Results) > 0 { obj["results"] = valuesToJSON(e.Results) }
    // Enrich call expressions with simple type signature hints for downstream phases.
    if e.Op == "call" {
        if len(e.Args) > 0 {
            ats := make([]any, 0, len(e.Args))
            for _, a := range e.Args { ats = append(ats, a.Type) }
            obj["argTypes"] = ats
        } else {
            obj["argTypes"] = []any{}
        }
        if len(e.Results) > 0 {
            rt := make([]any, 0, len(e.Results))
            for _, r := range e.Results { rt = append(rt, r.Type) }
            obj["retTypes"] = rt
        } else if e.Result != nil {
            obj["retTypes"] = []any{e.Result.Type}
        } else { obj["retTypes"] = []any{} }
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
        } else if len(e.Results) > 0 {
            for _, r := range e.Results { rs = append(rs, r.Type) }
        } else if e.Result != nil { rs = append(rs, e.Result.Type) }
        sig["params"] = ps
        sig["results"] = rs
        obj["sig"] = sig
    }
    return obj
}
