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
        case "Optional":
            if len(parts) != 1 { return nil, fmt.Errorf("Optional requires one type arg") }
            a, err := Parse(parts[0])
            if err != nil { return nil, err }
            return Optional{Inner: a}, nil
        case "Union":
            if len(parts) < 2 { return nil, fmt.Errorf("Union requires two or more type args") }
            alts := make([]Type, 0, len(parts))
            seen := map[string]struct{}{}
            for _, p := range parts {
                t, err := Parse(p)
                if err != nil { return nil, err }
                key := t.String()
                if _, ok := seen[key]; !ok { alts = append(alts, t); seen[key] = struct{}{} }
            }
            return Union{Alts: alts}, nil
        default:
            // For unknown bases, accept one or more type arguments to support nested
            // arity analysis (e.g., Owned<int,string>). Parse each arg recursively.
            if len(parts) == 0 { return Generic{Name: base}, nil }
            args := make([]Type, 0, len(parts))
            for _, p := range parts {
                a, err := Parse(p)
                if err != nil { return nil, err }
                args = append(args, a)
            }
            return Generic{Name: base, Args: args}, nil
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
    // default: named type without args (simple named/generic placeholder)
    return Generic{Name: s}, nil
}

// splitTopAll splits on commas at top level, respecting nesting of '<>', '{}', '()' and string quotes.
// splitTop splits a comma-separated list without breaking nested <...> groups.
// splitTop removed in favor of splitAllTop from fromast.go
