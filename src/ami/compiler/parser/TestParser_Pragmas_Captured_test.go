package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Pragmas_Captured(t *testing.T) {
    src := "package app\n#pragma test:case basic\n#pragma lint:disable W_IDENT_UNDERSCORE count=2 msg=\"x\"\nfunc F(){}\n"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Pragmas) != 2 { t.Fatalf("want 2 pragmas, got %d", len(file.Pragmas)) }
    if p := file.Pragmas[0]; p.Domain != "test" || p.Key != "case" || p.Value != "basic" { t.Fatalf("p0: %+v", p) }
    if p := file.Pragmas[1]; p.Domain != "lint" || p.Key != "disable" { t.Fatalf("p1: %+v", p) }
    if file.Pragmas[1].Params["count"] != "2" || file.Pragmas[1].Params["msg"] != "x" { t.Fatalf("params: %+v", file.Pragmas[1].Params) }
    if file.Pragmas[0].Pos.Line != 2 || file.Pragmas[1].Pos.Line != 3 { t.Fatalf("pragma positions wrong: %+v", file.Pragmas) }
}

