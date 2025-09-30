package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestIR_TopLevel_Config_FromPragmas(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\n#pragma concurrency level=8\n#pragma concurrency:schedule fair\n#pragma backpressure policy=block\n#pragma telemetry enabled=true\nfunc F(){}\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    path := filepath.Join("build", "debug", "ir", "app", "u.ir.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if obj["concurrency"].(float64) != 8 { t.Fatalf("concurrency: %v", obj["concurrency"]) }
    if obj["backpressurePolicy"].(string) != "block" { t.Fatalf("backpressurePolicy: %v", obj["backpressurePolicy"]) }
    if tv, ok := obj["telemetryEnabled"].(bool); !ok || !tv { t.Fatalf("telemetryEnabled: %v", obj["telemetryEnabled"]) }
    if obj["schedule"].(string) != "fair" { t.Fatalf("schedule: %v", obj["schedule"]) }
}
