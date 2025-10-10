package driver

import (
    "encoding/json"
    "path/filepath"
    "testing"

    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// Verify that inline worker literals are normalized to generated names in pipelines debug.
func TestPipelinesDebug_InlineWorker_Normalized(t *testing.T) {
    f := &ast.File{PackageName: "app"}
    // pipeline P { Transform func(e Event<int>) (Event<int>, error); }
    pd := &ast.PipelineDecl{Name: "P"}
    st := &ast.StepStmt{Name: "Transform", Args: []ast.Arg{{Text: "func(e Event<int>) (Event<int>, error)"}}}
    pd.Stmts = append(pd.Stmts, st)
    f.Decls = append(f.Decls, pd)
    path, err := writePipelinesDebug("app", "u", f)
    if err != nil { t.Fatalf("write pipelines: %v", err) }
    b := mustRead(t, path)
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    pipes := obj["pipelines"].([]any)
    if len(pipes) == 0 { t.Fatalf("no pipelines") }
    steps := pipes[0].(map[string]any)["steps"].([]any)
    if len(steps) == 0 { t.Fatalf("no steps") }
    args := steps[0].(map[string]any)["args"].([]any)
    if len(args) == 0 { t.Fatalf("no args") }
    first := args[0].(string)
    if !(len(first) >= len("InlineWorker_") && first[:len("InlineWorker_")] == "InlineWorker_") {
        t.Fatalf("arg not normalized: %q", first)
    }
    // also verify workers.symbols and workers_impl list include generated worker later in compile path; here we only check pipelines.
    _ = filepath.Base(path)
}

