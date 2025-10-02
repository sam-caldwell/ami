package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

func posPtr(p source.Position) *dbgPos {
    if p.Line == 0 && p.Column == 0 && p.Offset == 0 { return nil }
    return &dbgPos{Line: p.Line, Column: p.Column, Offset: p.Offset}
}

