package parser

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// ParseError is a structured parse error with a stable code and position.
// Code values are stable for testing; for now, all parser-generated errors use "E_PARSE".
type ParseError struct {
    Pos  source.Position
    Code string
    Msg  string
}

func (e ParseError) Error() string { return e.Msg }

