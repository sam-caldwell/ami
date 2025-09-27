package token

// LookupKeyword returns the keyword Kind for the given lower-case lexeme.
// Returns (kind, true) when the lexeme is a reserved keyword; otherwise (Ident, false).
func LookupKeyword(lexeme string) (Kind, bool) {
    if k, ok := Keywords[lexeme]; ok {
        return k, true
    }
    return Ident, false
}

