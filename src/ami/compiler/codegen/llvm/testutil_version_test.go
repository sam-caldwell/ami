package llvm

import "strconv"

// parseClangMajor extracts the leading major version from a clang --version line.
// Returns 0 if it cannot parse a number.
func parseClangMajor(s string) int {
    for i := 0; i < len(s); i++ {
        if s[i] >= '0' && s[i] <= '9' {
            j := i
            for j < len(s) && s[j] >= '0' && s[j] <= '9' { j++ }
            if n, err := strconv.Atoi(s[i:j]); err == nil { return n }
            break
        }
    }
    return 0
}

