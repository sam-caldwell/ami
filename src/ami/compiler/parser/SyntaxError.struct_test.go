package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestSyntaxError_Struct_Fields(t *testing.T) {
    se := SyntaxError{Msg: "oops", Pos: source.Position{Line: 1, Column: 2, Offset: 3}}
    if se.Msg != "oops" {
        t.Fatalf("Msg: want %q, got %q", "oops", se.Msg)
    }
    if se.Pos.Line != 1 || se.Pos.Column != 2 || se.Pos.Offset != 3 {
        t.Fatalf("Pos: want {1 2 3}, got %+v", se.Pos)
    }
}
