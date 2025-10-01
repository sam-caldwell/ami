package parser

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// SyntaxError represents a parser error with position information.
// It implements error and carries the current token position.
type SyntaxError struct {
    Msg string
    Pos source.Position
}

