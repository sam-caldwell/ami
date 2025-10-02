package sem

func elemCompatible(a, b string) bool {
    ea := innerGeneric(a)
    eb := innerGeneric(b)
    if isTypeVar(ea) || isTypeVar(eb) { return true }
    return typesCompatible(ea, eb)
}

