package driver

// ifReturn captures parsed if/else return forms.
type ifReturn struct {
    lhs string
    op  string
    rhs string
    thenIsEv bool
    thenLit  string
    elseIsEv bool
    elseLit  string
}

