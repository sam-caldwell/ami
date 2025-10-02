package driver

// getInt parses a positive integer from a string (digits only); returns 0 on invalid input.
func getInt(s string) int {
    n := 0
    for i := 0; i < len(s); i++ {
        if s[i] >= '0' && s[i] <= '9' { n = n*10 + int(s[i]-'0') } else { return 0 }
    }
    return n
}

