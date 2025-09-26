package ast

// Comment attaches source comment text to a position in the source.
type Comment struct {
    Text string
    Pos  Position
}

