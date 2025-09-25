package diag

import (
	srcset "github.com/sam-caldwell/ami/src/ami/compiler/source"
	sch "github.com/sam-caldwell/ami/src/schemas"
)

type Level string

const (
	Info  Level = "info"
	Warn  Level = "warn"
	Error Level = "error"
)

// Diagnostic is the compiler-internal diagnostic format with optional position.
type Diagnostic struct {
	Level   Level
	Code    string
	Message string
	Package string
	File    string
	Pos     *srcset.Position
	EndPos  *srcset.Position
	Data    map[string]interface{}
}

// ToSchema converts the diagnostic to the public schemas.DiagV1 type.
func (d Diagnostic) ToSchema() sch.DiagV1 {
	var pos *sch.Position
	var end *sch.Position
	if d.Pos != nil {
		pos = &sch.Position{Line: d.Pos.Line, Column: d.Pos.Column, Offset: d.Pos.Offset}
	}
	if d.EndPos != nil {
		end = &sch.Position{Line: d.EndPos.Line, Column: d.EndPos.Column, Offset: d.EndPos.Offset}
	}
	return sch.DiagV1{
		Schema:  "diag.v1",
		Level:   string(d.Level),
		Code:    d.Code,
		Message: d.Message,
		Package: d.Package,
		File:    d.File,
		Pos:     pos,
		EndPos:  end,
		Data:    d.Data,
	}
}
