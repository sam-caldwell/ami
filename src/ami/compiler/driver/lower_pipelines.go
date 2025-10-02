package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerPipelines extracts Collect/merge attributes into IR pipelines for a single AST file.
func lowerPipelines(f *ast.File) []ir.Pipeline {
    var out []ir.Pipeline
    if f == nil { return out }
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        pl := ir.Pipeline{Name: pd.Name}
        for _, s := range pd.Stmts {
            st, ok := s.(*ast.StepStmt)
            if !ok || st.Name != "Collect" { continue }
            if mp := toMergePlan(st); mp != nil { pl.Collect = append(pl.Collect, ir.CollectSpec{Step: st.Name, Merge: mp}) }
        }
        if len(pl.Collect) > 0 { out = append(out, pl) }
    }
    return out
}

