package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestSyntaxError_Methods(t *testing.T) {
    se := SyntaxError{Msg: "m", Pos: source.Position{Line: 1, Column: 2, Offset: 3}}
    if se.Error() != "m" { t.Fatal("Error()") }
    p := se.Position()
    if p.Line != 1 || p.Column != 2 || p.Offset != 3 { t.Fatalf("Position(): %+v", p) }
}

