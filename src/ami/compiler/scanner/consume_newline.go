package scanner

// consumeNewline advances the scanner by one byte when the current character
// is a line feed, updating position counters accordingly.
func consumeNewline(is *Scanner) {
    if is.off < len(is.src) && is.src[is.off] == '\n' {
        is.off++
        is.line++
        is.column = 1
    }
}

