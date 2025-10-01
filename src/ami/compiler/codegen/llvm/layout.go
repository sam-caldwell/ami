package llvm

import (
    "sort"
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
)

// slotsOf returns the number of 8-byte slots needed to store type t.
// Primitive and handle-like types occupy one slot; Structs sum the slots of
// their fields; Optional takes the inner; Union takes the max.
func slotsOf(t types.Type) int64 {
    switch v := t.(type) {
    case types.Primitive:
        return 1
    case types.Optional:
        return slotsOf(v.Inner)
    case types.Union:
        var mx int64
        for _, a := range v.Alts { if s := slotsOf(a); s > mx { mx = s } }
        if mx == 0 { mx = 1 }
        return mx
    case types.Struct:
        // Stable order by field name
        names := make([]string, 0, len(v.Fields))
        for k := range v.Fields { names = append(names, k) }
        sort.Strings(names)
        var sum int64
        for _, n := range names { sum += slotsOf(v.Fields[n]) }
        if sum == 0 { sum = 1 }
        return sum
    case types.Generic:
        // Named non-primitive types (Time, Duration, Owned, string, containers, Event, map, etc.) → 1 slot
        return 1
    default:
        return 1
    }
}

// fieldOffsetSlots computes the slot offset for a dotted path within a Struct type.
// Returns the offset in 8-byte slots and the leaf type.
func fieldOffsetSlots(root types.Type, path string) (int64, types.Type, bool) {
    if root == nil || path == "" { return 0, nil, false }
    segs := strings.Split(path, ".")
    cur := root
    var off int64
    for i, s := range segs {
        // unwrap Optional for traversal
        if o, ok := cur.(types.Optional); ok { cur = o.Inner }
        st, ok := cur.(types.Struct)
        if !ok { return 0, nil, false }
        // stable order fields with sizes
        names := make([]string, 0, len(st.Fields))
        for k := range st.Fields { names = append(names, k) }
        sort.Strings(names)
        var found bool
        for _, n := range names {
            if n == s {
                found = true
                if i == len(segs)-1 {
                    cur = st.Fields[n]
                } else {
                    cur = st.Fields[n]
                }
                break
            }
            off += slotsOf(st.Fields[n])
        }
        if !found { return 0, nil, false }
    }
    return off, cur, true
}
