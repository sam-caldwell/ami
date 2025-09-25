package metrics

import (
    "bytes"
    "encoding/json"
    "os"
    "strings"
    "testing"
    lg "github.com/sam-caldwell/ami/src/internal/logger"
)

func captureStdout(f func()) string {
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    defer func(){ os.Stdout = old }()
    f()
    w.Close()
    var buf bytes.Buffer
    _, _ = buf.ReadFrom(r)
    return buf.String()
}

func TestMetrics_Pipeline_EmitsDiagV1JSON(t *testing.T) {
    t.Cleanup(func(){ lg.Setup(false,false,false) })
    lg.Setup(true, true, false)
    out := captureStdout(func(){
        PipelineMetrics{Pipeline:"P", QueueDepth:7, Throughput:123.4, LatencyMs:20, Errors:1}.Emit()
    })
    line := strings.TrimSpace(out)
    var m map[string]interface{}
    if err := json.Unmarshal([]byte(line), &m); err != nil { t.Fatalf("unmarshal: %v; out=%q", err, out) }
    if m["schema"] != "diag.v1" || m["level"] != "info" || m["message"] != "metrics.pipeline" { t.Fatalf("unexpected header: %+v", m) }
    data, ok := m["data"].(map[string]interface{})
    if !ok { t.Fatalf("missing data object: %+v", m) }
    if data["pipeline"] != "P" || data["queueDepth"].(float64) != 7 || data["latencyMs"].(float64) != 20 || data["errors"].(float64) != 1 {
        t.Fatalf("missing fields in data: %+v", data)
    }
}

func TestMetrics_Node_EmitsDiagV1JSON(t *testing.T) {
    t.Cleanup(func(){ lg.Setup(false,false,false) })
    lg.Setup(true, true, false)
    out := captureStdout(func(){
        NodeMetrics{Pipeline:"P", Node:"Transform", QueueDepth:3, Throughput:42.0, LatencyMs:5, Errors:0}.Emit()
    })
    var m map[string]interface{}
    if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &m); err != nil { t.Fatalf("unmarshal: %v; out=%q", err, out) }
    data, ok := m["data"].(map[string]interface{})
    if !ok || data["node"] != "Transform" || data["pipeline"] != "P" { t.Fatalf("unexpected data: %+v", data) }
}
