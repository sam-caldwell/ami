package main

func equalSHA(a, b string) bool {
    if len(a) != len(b) { return false }
    var diff byte
    for i := 0; i < len(a); i++ { diff |= a[i] ^ b[i] }
    return diff == 0
}

