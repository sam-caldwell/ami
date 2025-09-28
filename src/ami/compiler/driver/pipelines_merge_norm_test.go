package driver

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestPipelinesDebug_MergeNormalization(t *testing.T) {
    pd := &ast.PipelineDecl{Name: "P"}
    // Collect with merge attrs
    st := &ast.StepStmt{Name: "Collect", Attrs: []ast.Attr{
        {Name: "merge.Buffer", Args: []ast.Arg{{Text: "4"}, {Text: "dropOldest"}}},
        {Name: "merge.Stable"},
        {Name: "merge.Sort", Args: []ast.Arg{{Text: "ts"}, {Text: "asc"}}},
    }}
    pd.Stmts = append(pd.Stmts, &ast.StepStmt{Name: "Ingress"})
    pd.Stmts = append(pd.Stmts, st)
    f := &ast.File{Decls: []ast.Decl{pd}}
    path, err := writePipelinesDebug("app", "u", f)
    if err != nil { t.Fatalf("write: %v", err) }
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read: %v", err) }
    var obj struct{
        Pipelines []struct{
            Steps []struct{
                MergeNorm any `json:"mergeNorm"`
            } `json:"steps"`
        } `json:"pipelines"`
    }
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if len(obj.Pipelines) == 0 || len(obj.Pipelines[0].Steps) < 2 { t.Fatalf("steps missing") }
    if obj.Pipelines[0].Steps[1].MergeNorm == nil { t.Fatalf("mergeNorm missing") }
}

