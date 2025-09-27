package logging

// Formatter converts a Record into a byte slice to write.
type Formatter interface {
    Format(r Record) []byte
}

