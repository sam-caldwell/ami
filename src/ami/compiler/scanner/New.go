package scanner

// New returns a new Scanner over the provided source string.
func New(src string) *Scanner {
    return &Scanner{
        src:    src,
        line:   1,
        column: 1,
    }
}

