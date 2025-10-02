package ir

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
        }
    }
    return out
}

