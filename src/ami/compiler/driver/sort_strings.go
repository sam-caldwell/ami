package driver

// sortStrings sorts a slice of strings in-place (small n typical).
func sortStrings(a []string) {
    for i := 1; i < len(a); i++ {
        j := i
        for j > 0 && a[j] < a[j-1] {
            a[j], a[j-1] = a[j-1], a[j]
            j--
        }
    }
}

