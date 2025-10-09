package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify that gpu(...) attribute list forms for grid/tpg are parsed and encoded in IR (2D/3D cases).
func TestGPUBlock_Dims_IR_Encoding(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    src := "package app\n" +
        "func W(){\n" +
        "  gpu(family=\"opencl\", name=\"k\", grid=[16,8,4], tpg=[8,4,2]){\n" +
        "    __kernel void k(__global long* out, unsigned int n){ size_t i=get_global_id(0); if(i<n) out[i] = (long)i * 3; }\n" +
        "  }\n" +
        "}\n"
    fs.AddFile("u.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ir := filepath.Join("build", "debug", "ir", "app", "u.ir.json")
    b, err := os.ReadFile(ir)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("unmarshal: %v", err) }
    fns, _ := obj["functions"].([]any)
    if len(fns) == 0 { t.Fatalf("no functions in IR") }
    fm := fns[0].(map[string]any)
    gbl, _ := fm["gpuBlocks"].([]any)
    if len(gbl) == 0 { t.Fatalf("no gpuBlocks in IR") }
    gb := gbl[0].(map[string]any)
    grid, _ := gb["grid"].([]any)
    if len(grid) != 3 { t.Fatalf("grid length: %d", len(grid)) }
    if int(grid[0].(float64)) != 16 || int(grid[1].(float64)) != 8 || int(grid[2].(float64)) != 4 {
        t.Fatalf("grid mismatch: %#v", grid)
    }
    tpg, _ := gb["tpg"].([]any)
    if len(tpg) != 3 { t.Fatalf("tpg length: %d", len(tpg)) }
    if int(tpg[0].(float64)) != 8 || int(tpg[1].(float64)) != 4 || int(tpg[2].(float64)) != 2 {
        t.Fatalf("tpg mismatch: %#v", tpg)
    }
}

