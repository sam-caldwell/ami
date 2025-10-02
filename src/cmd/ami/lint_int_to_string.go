package main

func intToString(i int) string {
    // fast small-int conversion without fmt
    if i == 0 { return "0" }
    neg := i < 0
    if neg { i = -i }
    var b [20]byte
    p := len(b)
    for i > 0 {
        p--
        b[p] = byte('0' + (i % 10))
        i /= 10
    }
    if neg { p--; b[p] = '-' }
    return string(b[p:])
}

