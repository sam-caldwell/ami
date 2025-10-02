package merge

func toKey(v any) string {
    switch x := v.(type) {
    case string: return x
    case int: return itoa(int64(x))
    case int64: return itoa(x)
    case float64: return ftoa(x)
    case bool: if x { return "true" } else { return "false" }
    default: return "" // non-deterministic; avoid map/array stringification here
    }
}

