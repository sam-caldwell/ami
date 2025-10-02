package driver

func trimQuotes(s string) string {
    if l := len(s); l >= 2 {
        if (s[0] == '"' && s[l-1] == '"') || (s[0] == '\'' && s[l-1] == '\'') { return s[1:l-1] }
    }
    return s
}

