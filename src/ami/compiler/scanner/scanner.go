package scanner

// Scanner tokenizes AMI source text into a stream of tokens.
// It tracks byte offset, line, and column for precise diagnostics.
type Scanner struct {
    src     string
    off     int
    line    int
    column  int
    pending []Comment
}

