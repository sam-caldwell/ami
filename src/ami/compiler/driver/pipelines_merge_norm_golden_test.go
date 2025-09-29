package driver

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// Verify merge normalization content exactly for a typical combination.
func TestPipelinesDebug_MergeNormalization_Golden(t *testing.T) {
    pd := &ast.PipelineDecl{Name: "P"}
    st := &ast.StepStmt{Name: "Collect", Attrs: []ast.Attr{
        {Name: "merge.Buffer", Args: []ast.Arg{{Text: "4"}, {Text: "dropOldest"}}},
        {Name: "merge.Stable"},
        {Name: "merge.Sort", Args: []ast.Arg{{Text: "ts"}, {Text: "asc"}}},
    }}
    pd.Stmts = append(pd.Stmts, &ast.StepStmt{Name: "Ingress"})
    pd.Stmts = append(pd.Stmts, st)
    pd.Stmts = append(pd.Stmts, &ast.StepStmt{Name: "Egress"})
    f := &ast.File{Decls: []ast.Decl{pd}}

    path, err := writePipelinesDebug("app", "u", f)
    if err != nil { t.Fatalf("write: %v", err) }
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read: %v", err) }

    var obj struct{
        Pipelines []struct{
            Steps []struct{
                MergeNorm struct{
                    Buffer *struct{
                        Capacity int    `json:"capacity"`
                        Policy   string `json:"policy"`
                    } `json:"buffer"`
                    Stable bool `json:"stable"`
                    Sort   []struct{
                        Field string `json:"field"`
                        Order string `json:"order"`
                    } `json:"sort"`
                } `json:"mergeNorm"`
            } `json:"steps"`
        } `json:"pipelines"`
    }
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if len(obj.Pipelines) != 1 { t.Fatalf("expected 1 pipeline, got %d", len(obj.Pipelines)) }
    steps := obj.Pipelines[0].Steps
    if len(steps) < 2 { t.Fatalf("expected at least 2 steps") }
    mn := steps[1].MergeNorm

    if mn.Buffer == nil || mn.Buffer.Capacity != 4 || mn.Buffer.Policy != "dropOldest" {
        t.Fatalf("buffer norm mismatch: %+v", mn.Buffer)
    }
    if !mn.Stable { t.Fatalf("stable expected true") }
    if len(mn.Sort) != 1 || mn.Sort[0].Field != "ts" || mn.Sort[0].Order != "asc" {
        t.Fatalf("sort norm mismatch: %+v", mn.Sort)
    }
}

