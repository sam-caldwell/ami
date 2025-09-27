package types

// Kind enumerates core builtin kinds.
type Kind int

const (
    Invalid Kind = iota
    Bool
    Int
    Int64
    Float64
    String
    // Additional primitives can be added as needed.
)

func (k Kind) String() string {
    switch k {
    case Bool: return "bool"
    case Int: return "int"
    case Int64: return "int64"
    case Float64: return "float64"
    case String: return "string"
    default: return "invalid"
    }
}

