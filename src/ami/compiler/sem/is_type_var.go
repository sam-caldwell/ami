package sem

func isTypeVar(s string) bool {
    if s == "any" || s == "" { return true }
    // consider single-letter ASCII uppercase as a type variable (T/U/E/etc.)
    return len(s) == 1 && s[0] >= 'A' && s[0] <= 'Z'
}

