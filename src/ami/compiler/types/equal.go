package types

// Equal reports structural equality of two types, ignoring Union alternative
// order. Struct field order is normalized by String() so direct comparison of
// keys with recursive Equal suffices.
func Equal(a, b Type) bool {
    switch av := a.(type) {
    case Primitive:
        bv, ok := b.(Primitive); return ok && av.K == bv.K
    case Named:
        bv, ok := b.(Named); return ok && av.Name == bv.Name
    case Generic:
        bv, ok := b.(Generic); if !ok { return false }
        if av.Name != bv.Name || len(av.Args) != len(bv.Args) { return false }
        // Special-case Union represented as Generic shouldn't occur; Union has its own type.
        for i := range av.Args { if !Equal(av.Args[i], bv.Args[i]) { return false } }
        return true
    case Struct:
        bv, ok := b.(Struct); if !ok { return false }
        if len(av.Fields) != len(bv.Fields) { return false }
        for k, v := range av.Fields {
            vv, ok := bv.Fields[k]; if !ok { return false }
            if !Equal(v, vv) { return false }
        }
        return true
    case Optional:
        bv, ok := b.(Optional); return ok && Equal(av.Inner, bv.Inner)
    case Union:
        bv, ok := b.(Union); if !ok { return false }
        if len(av.Alts) != len(bv.Alts) { return false }
        // Compare as sets using string keys built from structural String forms.
        set := make(map[string]struct{}, len(av.Alts))
        for _, x := range av.Alts { set[x.String()] = struct{}{} }
        for _, y := range bv.Alts { if _, ok := set[y.String()]; !ok { return false } }
        return true
    default:
        // Unknown concrete type; fall back to string comparison as last resort
        return a != nil && b != nil && a.String() == b.String()
    }
}

