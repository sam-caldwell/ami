package main

func atoiSafe(s string) int {
    // simple unsigned parse; non-numeric returns 0
    n := 0
    for i := 0; i < len(s); i++ {
        c := s[i]
        if c < '0' || c > '9' { break }
        n = n*10 + int(c-'0')
        if n > 1_000_000_000 { break }
    }
    return n
}

