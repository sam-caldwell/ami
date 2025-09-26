package tester

import "encoding/json"

func asInt(v any) (int, bool) {
    switch t := v.(type) {
    case float64:
        return int(t), true
    case json.Number:
        if i, err := t.Int64(); err == nil {
            return int(i), true
        }
    case int:
        return t, true
    }
    return 0, false
}

