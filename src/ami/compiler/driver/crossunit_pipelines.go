package driver

import (
    "time"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// analyzeDuplicatePipelinesAcrossUnits emits E_DUP_PIPELINE when the same pipeline name
// appears in multiple units within a package. It reports the later occurrences and embeds
// the previous occurrence position in diagnostic data.
func analyzeDuplicatePipelinesAcrossUnits(units []struct{ file *source.File; ast *ast.File; unit string }) []diag.Record {
    var out []diag.Record
    if len(units) == 0 { return out }
    now := time.Unix(0, 0).UTC()
    type pos struct{ file string; p diag.Position }
    seen := map[string]pos{}
    for _, u := range units {
        if u.ast == nil { continue }
        for _, d := range u.ast.Decls {
            pd, ok := d.(*ast.PipelineDecl)
            if !ok { continue }
            key := pd.Name
            cur := diag.Position{Line: pd.NamePos.Line, Column: pd.NamePos.Column, Offset: pd.NamePos.Offset}
            if prev, ok := seen[key]; ok {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DUP_PIPELINE", Message: "duplicate pipeline name: " + key, File: u.file.Name, Pos: &cur, Data: map[string]any{"previous": prev.p, "previousFile": prev.file}})
            } else {
                seen[key] = pos{file: u.file.Name, p: cur}
            }
        }
    }
    return out
}

