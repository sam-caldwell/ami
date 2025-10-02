package tester

func asInt(v any) (int, bool) {
    switch n := v.(type) {
    case int:
        return n, true
    case int32:
        return int(n), true
    case int64:
        return int(n), true
    case float64:
        return int(n), true
    case float32:
        return int(n), true
    default:
        return 0, false
    }
}

