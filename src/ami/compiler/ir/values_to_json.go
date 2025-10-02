package ir

func valuesToJSON(vs []Value) []any {
    out := make([]any, 0, len(vs))
    for _, v := range vs { out = append(out, map[string]any{"id": v.ID, "type": v.Type}) }
    return out
}

