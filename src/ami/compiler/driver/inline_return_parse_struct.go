package driver

// returnParse captures parsed return expression form.
type returnParse struct {
    kind retKind
    lit  string
    lhs  string
    rhs  string
    op   string // one of +,-,*,/,% or a comparison op
    lhsIsEv bool
    rhsIsEv bool
}

