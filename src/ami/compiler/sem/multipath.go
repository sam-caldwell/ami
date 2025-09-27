package sem

import (
    "time"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeMultiPath validates MultiPath usage and merge.* attributes on Collect steps.
// - edge.MultiPath/MultiPath only valid on Collect → E_MP_ONLY_COLLECT
// - merge.* unknown → E_MERGE_ATTR_UNKNOWN
// - merge.* invalid arity → E_MERGE_ATTR_ARGS
// - merge.Sort without a field → W_MERGE_SORT_NO_FIELD
func AnalyzeMultiPath(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // allowed merge attributes and arity constraints (min,max; -1 means unbounded)
    type ar struct{ min, max int }
    merges := map[string]ar{
        "merge.Sort":       {1, 2},
        "merge.Stable":     {0, 0},
        "merge.Key":        {1, 1},
        "merge.Dedup":      {0, 1},
        "merge.Window":     {1, 1},
        "merge.Watermark":  {2, 2},
        "merge.Timeout":    {1, 1},
        "merge.Buffer":     {1, 2},
        "merge.PartitionBy":{1, 1},
    }
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        for _, s := range pd.Stmts {
            st, ok := s.(*ast.StepStmt)
            if !ok { continue }
            // detect multipath on non-Collect
            for _, at := range st.Attrs {
                if at.Name == "edge.MultiPath" || at.Name == "MultiPath" {
                    if st.Name != "Collect" {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MP_ONLY_COLLECT", Message: "edge.MultiPath only valid on Collect nodes", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                    }
                }
            }
            if st.Name != "Collect" { continue }
            // validate merge.* attributes
            for _, at := range st.Attrs {
                if strings.HasPrefix(at.Name, "merge.") {
                    if rng, ok := merges[at.Name]; ok {
                        argc := len(at.Args)
                        if at.Name == "merge.Sort" && argc == 0 {
                            out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_SORT_NO_FIELD", Message: "merge.Sort requires a field", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                            continue
                        }
                        if argc < rng.min || (rng.max >= 0 && argc > rng.max) {
                            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_ARGS", Message: at.Name + ": invalid number of arguments", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                            continue
                        }
                    } else {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_UNKNOWN", Message: "unknown merge attribute: " + at.Name, Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                    }
                }
            }
        }
    }
    return out
}
