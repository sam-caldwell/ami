package enum

// MustParse panics if name is invalid.
func MustParse(d Descriptor, s string) int {
    v, err := Parse(d, s)
    if err != nil { panic(err) }
    return v
}

