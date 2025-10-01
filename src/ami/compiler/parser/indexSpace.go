package parser

// indexSpace returns the index of the first ASCII space or tab, or -1.
func indexSpace(s string) int {
    for i := 0; i < len(s); i++ {
        if s[i] == ' ' || s[i] == '\t' {
            return i
        }
    }
    return -1
}

