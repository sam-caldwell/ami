package types

import (
    "strings"
    "unicode"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// FromAST maps an AST type reference (currently represented as a string in the AST)
// into a concrete types.Type. Supports primitives, generics, pointer/slice forms,
// and container generics: Event<T>, Error<E>, Owned<T>, slice<T>, set<T>, map<K,V>.
func FromAST(t string) Type {
    s := strings.TrimSpace(t)
    if s == "" { return Named{Name: ""} }
    // pointer form
    if strings.HasPrefix(s, "*") {
        return Pointer{Elem: FromAST(strings.TrimSpace(s[1:]))}
    }
    // bracket slice form
    if strings.HasPrefix(s, "[]") {
        return Slice{Elem: FromAST(strings.TrimSpace(s[2:]))}
    }
    // map generic
    if strings.HasPrefix(s, "map<") && strings.HasSuffix(s, ">") {
        inner := s[len("map<") : len(s)-1]
        k, v := splitTop(inner)
        return Map{Key: FromAST(k), Val: FromAST(v)}
    }
    // set generic
    if strings.HasPrefix(s, "set<") && strings.HasSuffix(s, ">") {
        inner := s[len("set<") : len(s)-1]
        return Set{Elem: FromAST(inner)}
    }
    // slice generic
    if strings.HasPrefix(s, "slice<") && strings.HasSuffix(s, ">") {
        inner := s[len("slice<") : len(s)-1]
        return SliceTy{Elem: FromAST(inner)}
    }
    // general generic: Name<...>
    if i := strings.IndexByte(s, '<'); i >= 0 && strings.HasSuffix(s, ">") {
        name := s[:i]
        inner := s[i+1 : len(s)-1]
        parts := splitAllTop(inner)
        args := make([]Type, 0, len(parts))
        for _, p := range parts { args = append(args, FromAST(p)) }
        return Generic{Name: name, Args: args}
    }
    // primitives by name
    switch s {
    case "bool": return Primitive{K: Bool}
    case "int": return Primitive{K: Int}
    case "int64": return Primitive{K: Int64}
    case "float64": return Primitive{K: Float64}
    case "string": return Primitive{K: String}
    }
    // single-letter type variable or user-defined named type
    if isTypeVarName(s) { return Named{Name: s} }
    return Named{Name: s}
}

// BuildFunction constructs a Function type from a FuncDecl's parameter/result type strings.
func BuildFunction(fn *ast.FuncDecl) Function {
    var ps, rs []Type
    if fn != nil {
        for _, p := range fn.Params { ps = append(ps, FromAST(p.Type)) }
        for _, r := range fn.Results { rs = append(rs, FromAST(r.Type)) }
    }
    return Function{Params: ps, Results: rs}
}

// splitTop splits "A,B" at top-level, respecting nested generics.
func splitTop(s string) (string, string) {
    depth := 0
    for i := 0; i < len(s); i++ {
        switch s[i] {
        case '<': depth++
        case '>': if depth > 0 { depth-- }
        case ',':
            if depth == 0 {
                return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:])
            }
        }
    }
    return strings.TrimSpace(s), ""
}

func splitAllTop(s string) []string {
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
                start = i+1
            }
        }
    }
    out = append(out, strings.TrimSpace(s[start:]))
    return out
}

func isTypeVarName(s string) bool {
    if len(s) == 1 {
        r := rune(s[0])
        return unicode.IsUpper(r)
    }
    return false
}

