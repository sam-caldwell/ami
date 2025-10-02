package driver

import (
    "fmt"
    "strconv"
)

func parseInt(text string) (int, error) {
    // strip optional sign
    if len(text) == 0 { return 0, fmt.Errorf("empty") }
    neg := false
    if text[0] == '-' { neg = true; text = text[1:] }
    base := 10
    if len(text) > 2 && text[0] == '0' {
        switch text[1] {
        case 'x', 'X': base = 16; text = text[2:]
        case 'b', 'B': base = 2;  text = text[2:]
        case 'o', 'O': base = 8;  text = text[2:]
        }
    }
    // remove underscores if any (future-proof)
    clean := make([]rune, 0, len(text))
    for _, r := range text { if r != '_' { clean = append(clean, r) } }
    n, err := strconv.ParseInt(string(clean), base, 64)
    if err != nil { return 0, err }
    if neg { n = -n }
    return int(n), nil
}

