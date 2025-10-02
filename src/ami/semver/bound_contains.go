package semver

// Contains reports whether v lies within bound b.
func Contains(b Bound, v Version) bool {
    // lower check
    cmp := Compare(v, b.Lower)
    if cmp < 0 || (cmp == 0 && !b.LowerInclusive) { return false }
    if b.Upper == nil { return true }
    cmp = Compare(v, *b.Upper)
    if cmp > 0 || (cmp == 0 && !b.UpperInclusive) { return false }
    return true
}

