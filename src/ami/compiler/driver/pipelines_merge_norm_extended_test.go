package driver

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestPipelinesDebug_MergeNormalization_Extended(t *testing.T) {
    pd := &ast.PipelineDecl{Name: "P"}
    st := &ast.StepStmt{Name: "Collect", Attrs: []ast.Attr{
        {Name: "merge.Key", Args: []ast.Arg{{Text: "id"}}},
        {Name: "merge.PartitionBy", Args: []ast.Arg{{Text: "pid"}}},
        {Name: "merge.Timeout", Args: []ast.Arg{{Text: "250"}}},
        {Name: "merge.Window", Args: []ast.Arg{{Text: "32"}}},
        {Name: "merge.Watermark", Args: []ast.Arg{{Text: "ts"}, {Text: "100ms"}}},
        {Name: "merge.Dedup", Args: []ast.Arg{{Text: "id"}}},
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
                MergeNorm struct{
                    Key         string `json:"key"`
                    PartitionBy string `json:"partitionBy"`
                    TimeoutMs   int    `json:"timeoutMs"`
                    Window      int    `json:"window"`
                    Watermark   *struct{
                        Field    string `json:"field"`
                        Lateness string `json:"lateness"`
                    } `json:"watermark"`
                    Dedup string `json:"dedup"`
                } `json:"mergeNorm"`
            } `json:"steps"`
        } `json:"pipelines"`
    }
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    mn := obj.Pipelines[0].Steps[1].MergeNorm
    if mn.Key != "id" || mn.PartitionBy != "pid" { t.Fatalf("key/partition mismatch: %+v", mn) }
    if mn.TimeoutMs != 250 || mn.Window != 32 { t.Fatalf("timeout/window mismatch: %+v", mn) }
    if mn.Watermark == nil || mn.Watermark.Field != "ts" || mn.Watermark.Lateness != "100ms" {
        t.Fatalf("watermark mismatch: %+v", mn.Watermark)
    }
    if mn.Dedup != "id" { t.Fatalf("dedup mismatch: %q", mn.Dedup) }
}

