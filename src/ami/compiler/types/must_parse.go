package types

// MustParse is like Parse but panics on error. Intended for tests/tooling.
func MustParse(s string) Type {
    t, err := Parse(s)
    if err != nil { panic(err) }
    return t
}

