package driver

// lastSlash returns the last index of '/' or -1
func lastSlash(s string) int {
    for i := len(s) - 1; i >= 0; i-- {
        if s[i] == '/' { return i }
    }
    return -1
}

