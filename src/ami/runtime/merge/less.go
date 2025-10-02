package merge

func less(a, b item, p Plan) bool {
    for i, k := range p.Sort {
        av := a.keys[i]
        bv := b.keys[i]
        c := cmp(av, bv)
        if c == 0 { continue }
        if k.Order == "desc" { return c > 0 }
        return c < 0
    }
    // Secondary tiebreak: explicit Key when provided
    if p.Key != "" {
        if c := cmp(a.key, b.key); c != 0 { return c < 0 }
    }
    if p.Stable { return a.seq < b.seq }
    return false
}

