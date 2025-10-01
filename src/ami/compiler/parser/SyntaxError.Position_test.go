package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestSyntaxError_Position(t *testing.T) {
    se := SyntaxError{Pos: source.Position{Line: 10, Column: 20, Offset: 30}}
    p := se.Position()
    if p.Line != 10 || p.Column != 20 || p.Offset != 30 {
        t.Fatalf("Position(): want {10 20 30}, got %+v", p)
    }
}

