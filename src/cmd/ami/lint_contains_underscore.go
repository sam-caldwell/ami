package main

func containsUnderscore(s string) bool {
    for i := 0; i < len(s); i++ { if s[i] == '_' { return true } }
    return false
}

