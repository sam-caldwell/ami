package driver

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestPipelinesDebug_MergeNormalization_KV_Buffer(t *testing.T) {
    // Buffer with keyed args in any order; last-write-wins on duplicates
    pd := &ast.PipelineDecl{Name: "P"}
    st := &ast.StepStmt{Name: "Collect", Attrs: []ast.Attr{
        {Name: "merge.Buffer", Args: []ast.Arg{{Text: "policy=dropNewest"}, {Text: "capacity=8"}}},
        {Name: "merge.Buffer", Args: []ast.Arg{{Text: "capacity=4"}, {Text: "policy=dropOldest"}}},
    }}
    pd.Stmts = []ast.Stmt{&ast.StepStmt{Name: "Ingress"}, st, &ast.StepStmt{Name: "Egress"}}
    f := &ast.File{Decls: []ast.Decl{pd}}
    path, err := writePipelinesDebug("app", "u", f)
    if err != nil { t.Fatalf("write: %v", err) }
    b, _ := os.ReadFile(path)
    var obj struct{
        Pipelines []struct{
            Steps []struct{
                MergeNorm struct{
                    Buffer *struct{ Capacity int `json:"capacity"`; Policy string `json:"policy"` } `json:"buffer"`
                } `json:"mergeNorm"`
            } `json:"steps"`
        } `json:"pipelines"`
    }
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    buf := obj.Pipelines[0].Steps[1].MergeNorm.Buffer
    if buf == nil || buf.Capacity != 4 || buf.Policy != "dropOldest" {
        t.Fatalf("expected last buffer wins (cap=4, policy=dropOldest), got %+v", buf)
    }
}

