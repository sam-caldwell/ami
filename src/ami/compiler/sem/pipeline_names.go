package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzePipelineNames enforces uniqueness of pipeline names while allowing
// multiple entrypoints (multiple pipelines) per program.
func AnalyzePipelineNames(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    seen := map[string]source.Position{}
    for _, d := range f.Decls {
        if pd, ok := d.(*ast.PipelineDecl); ok {
            if prev, found := seen[pd.Name]; found {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PIPELINE_NAME_DUP", Message: "duplicate pipeline name: " + pd.Name, Pos: &diag.Position{Line: pd.NamePos.Line, Column: pd.NamePos.Column, Offset: pd.NamePos.Offset}, Data: map[string]any{"prevLine": prev.Line}})
            } else {
                seen[pd.Name] = source.Position{Line: pd.NamePos.Line, Column: pd.NamePos.Column, Offset: pd.NamePos.Offset}
            }
        }
    }
    return out
}
