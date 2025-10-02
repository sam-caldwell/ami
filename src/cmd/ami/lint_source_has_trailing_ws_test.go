package main

import "testing"

func Test_hasTrailingWhitespace(t *testing.T) {
    if hasTrailingWhitespace("") { t.Fatal("empty should be false") }
    if !hasTrailingWhitespace("abc ") { t.Fatal("space suffix true") }
    if !hasTrailingWhitespace("abc\t") { t.Fatal("tab suffix true") }
    if hasTrailingWhitespace(" abc") { t.Fatal("leading ws not trailing") }
}

