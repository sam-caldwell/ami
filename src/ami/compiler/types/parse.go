package types

import (
    "fmt"
    "strings"
)

// Parse converts a textual type (e.g., "Event<int>", "map<string,int>") into a Type.
// It supports primitives (bool,int,int64,float64,string), single-arg generics
// (slice<T>, set<T>, Event<T>, Error<E>), and two-arg map<K,V>.
// Unknown base names are returned as Generic{Name: base} with parsed args.
func Parse(s string) (Type, error) {
    s = strings.TrimSpace(s)
    if s == "" { return Primitive{K: Invalid}, fmt.Errorf("empty type") }
    // primitives
    switch s {
    case "bool": return Primitive{K: Bool}, nil
    case "int": return Primitive{K: Int}, nil
    case "int64": return Primitive{K: Int64}, nil
    case "float64": return Primitive{K: Float64}, nil
    case "string": return Primitive{K: String}, nil
    }
    // generic forms: name<...>
    if i := strings.IndexByte(s, '<'); i >= 0 && strings.HasSuffix(s, ">") {
        base := s[:i]
        inner := s[i+1:len(s)-1]
        // split by top-level commas (no nested generics beyond single level needed for now)
        parts := splitTop(inner)
        switch base {
        case "map":
            if len(parts) != 2 { return nil, fmt.Errorf("map requires two type args") }
            k, err := Parse(parts[0])
            if err != nil { return nil, err }
            v, err := Parse(parts[1])
            if err != nil { return nil, err }
            return Generic{Name: "map", Args: []Type{k, v}}, nil
        default:
            if len(parts) != 1 { return nil, fmt.Errorf("%s requires one type arg", base) }
            a, err := Parse(parts[0])
            if err != nil { return nil, err }
            return Generic{Name: base, Args: []Type{a}}, nil
        }
    }
    // default: named type without args
    return Generic{Name: s}, nil
}

// splitTop splits a comma-separated list without breaking nested <...> groups.
func splitTop(s string) []string {
    var out []string
    depth := 0
    start := 0
    for i := 0; i < len(s); i++ {
        switch s[i] {
        case '<': depth++
        case '>': if depth > 0 { depth-- }
        case ',':
            if depth == 0 {
                out = append(out, strings.TrimSpace(s[start:i]))
                start = i + 1
            }
        }
    }
    if start <= len(s) { out = append(out, strings.TrimSpace(s[start:])) }
    return out
}

