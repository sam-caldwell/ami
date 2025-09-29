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
        // split by top-level commas; respect '<>', '{}', '()' and quotes
        parts := splitTopAll(inner)
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
    // Struct{field:type,...}
    if strings.HasPrefix(s, "Struct{") && strings.HasSuffix(s, "}") {
        body := s[len("Struct{") : len(s)-1]
        fields := map[string]Type{}
        if strings.TrimSpace(body) != "" {
            parts := splitTopAll(body)
            for _, p := range parts {
                if i := strings.IndexByte(p, ':'); i > 0 {
                    name := strings.TrimSpace(p[:i])
                    tystr := strings.TrimSpace(p[i+1:])
                    ty, err := Parse(tystr)
                    if err != nil { return nil, err }
                    if name != "" { fields[name] = ty }
                } else {
                    return nil, fmt.Errorf("invalid struct field: %s", p)
                }
            }
        }
        return Struct{Fields: fields}, nil
    }
    // default: named type without args
    return Generic{Name: s}, nil
}

// splitTopAll splits on commas at top level, respecting nesting of '<>', '{}', '()' and string quotes.
func splitTopAll(s string) []string {
    var parts []string
    depthAngle, depthBrace, depthParen := 0, 0, 0
    inQuote := byte(0)
    last := 0
    for i := 0; i < len(s); i++ {
        c := s[i]
        if inQuote != 0 {
            if c == inQuote { inQuote = 0 }
            continue
        }
        switch c {
        case '\'', '"':
            inQuote = c
        case '<':
            depthAngle++
        case '>':
            if depthAngle > 0 { depthAngle-- }
        case '{':
            depthBrace++
        case '}':
            if depthBrace > 0 { depthBrace-- }
        case '(':
            depthParen++
        case ')':
            if depthParen > 0 { depthParen-- }
        case ',':
            if depthAngle == 0 && depthBrace == 0 && depthParen == 0 {
                parts = append(parts, strings.TrimSpace(s[last:i]))
                last = i + 1
            }
        }
    }
    tail := strings.TrimSpace(s[last:])
    if tail != "" { parts = append(parts, tail) }
    return parts
}

// splitTop splits a comma-separated list without breaking nested <...> groups.
// splitTop removed in favor of splitAllTop from fromast.go
