package sem

// isGenericArityMismatch compares two textual types and reports whether they
// refer to the same generic base but with a different number of type arguments.
func isGenericArityMismatch(expected, actual string) (bool, string, int, int) {
    eb, ea, eok := genericBaseAndArity(expected)
    ab, aa, aok := genericBaseAndArity(actual)
    if !eok || !aok { return false, "", 0, 0 }
    if eb != ab { return false, "", 0, 0 }
    if ea != aa { return true, eb, ea, aa }
    return false, eb, ea, aa
}

