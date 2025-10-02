package llvm

import (
    "sort"
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
)

// fieldOffsetSlots computes the slot offset for a dotted path within a Struct type.
// Returns the offset in 8-byte slots and the leaf type.
func fieldOffsetSlots(root types.Type, path string) (int64, types.Type, bool) {
    if root == nil || path == "" { return 0, nil, false }
    segs := strings.Split(path, ".")
    cur := root
    var off int64
    for _, s := range segs {
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
                cur = st.Fields[n]
                break
            }
            off += slotsOf(st.Fields[n])
        }
        if !found { return 0, nil, false }
    }
    return off, cur, true
}

