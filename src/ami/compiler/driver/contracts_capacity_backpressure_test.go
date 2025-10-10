package driver

import (
    "encoding/json"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestContractsDebug_IncludesCapacityAndBackpressure(t *testing.T) {
    code := "package app\n" +
        "pipeline P(){\n" +
        "  Collect edge.FIFO(min=1, max=4, backpressure=dropNewest)\n" +
        "}"
    f := &source.File{Name: "u.ami", Content: code}
    af := mustParse(t, f)
    path, err := writeContractsDebug("app", "u", af)
    if err != nil { t.Fatalf("contracts: %v", err) }
    b := mustRead(t, path)
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    pipes := obj["pipelines"].([]any)
    if len(pipes) == 0 { t.Fatalf("no pipelines") }
    steps := pipes[0].(map[string]any)["steps"].([]any)
    if len(steps) == 0 { t.Fatalf("no steps") }
    st := steps[0].(map[string]any)
    if st["minCapacity"].(float64) != 1 || st["maxCapacity"].(float64) != 4 || st["backpressure"].(string) != "dropNewest" {
        t.Fatalf("unexpected step fields: %+v", st)
    }
}
