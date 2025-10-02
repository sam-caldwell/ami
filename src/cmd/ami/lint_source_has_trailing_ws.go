package main

func hasTrailingWhitespace(s string) bool {
    if s == "" { return false }
    // spaces or tabs at end
    i := len(s) - 1
    if s[i] == ' ' || s[i] == '\t' { return true }
    return false
}

