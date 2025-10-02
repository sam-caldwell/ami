package types

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

