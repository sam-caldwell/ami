package types

import "strings"

// ResolveField attempts to resolve a dotted field path against a root type.
// If the root is Event<...>, resolution starts at the payload type. Optional
// unwraps for navigation but propagates to the result. Union resolves each
// alternative and returns a Union of resolved leaves if all alternatives
// resolve; otherwise resolution fails. Map/Set/Slice are terminal and cannot
// be navigated by named fields.
func ResolveField(root Type, path string) (Type, bool) {
    if root == nil { return nil, false }
    // If Event<...>, navigate into payload
    if g, ok := root.(Generic); ok && g.Name == "Event" && len(g.Args) == 1 {
        root = g.Args[0]
    }
    // split field path
    segs := []string{}
    if path != "" { segs = strings.Split(path, ".") }
    // propagate optional
    opt := false
    cur := root
    for _, s := range segs {
        // unwrap Optional for traversal
        if o, ok := cur.(Optional); ok {
            opt = true
            cur = o.Inner
        }
        // handle Union by resolving each alternative; require all alts resolve
        if u, ok := cur.(Union); ok {
            var leaves []Type
            for _, alt := range u.Alts {
                lt, ok := ResolveField(alt, s)
                if !ok { return nil, false }
                leaves = append(leaves, lt)
            }
            cur = Union{Alts: leaves}
            continue
        }
        // navigate Struct field
        if st, ok := cur.(Struct); ok {
            nt, ok := st.Fields[s]
            if !ok { return nil, false }
            cur = nt
            continue
        }
        // unsupported navigation target
        return nil, false
    }
    if opt { return Optional{Inner: cur}, true }
    return cur, true
}

// IsOrderable reports whether a type can be used for ordering (e.g., merge.Sort).
// Current rule: primitives are orderable; Optional<T> is orderable if T is orderable;
// Union is orderable only if all alternatives are orderable. Containers and Struct
// are not orderable.
func IsOrderable(t Type) bool {
    switch v := t.(type) {
    case Primitive:
        return true
    case Optional:
        return IsOrderable(v.Inner)
    case Union:
        if len(v.Alts) == 0 { return false }
        for _, a := range v.Alts { if !IsOrderable(a) { return false } }
        return true
    default:
        return false
    }
}

