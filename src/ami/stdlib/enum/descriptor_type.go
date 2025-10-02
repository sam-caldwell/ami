package enum

// Descriptor describes a generated enum: its type name and canonical member names
// in ordinal order (0..N-1). Names are case-sensitive and must be unique.
type Descriptor struct {
    Name  string
    Names []string
    idx   map[string]int
}

