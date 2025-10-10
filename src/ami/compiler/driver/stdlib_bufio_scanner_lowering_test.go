package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify bufio.Scanner signatures lower to EXPR call nodes with expected callees in IR debug JSON.
func TestStdlib_Bufio_Scanner_Lowering_Callees(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\n" +
        "import bufio\n" +
        "func F(){\n" +
        "  var s bufio.Scanner\n" +
        "  var ok bool\n" +
        "  ok = s.Scan()\n" +
        "  var txt string\n" +
        "  txt = s.Text()\n" +
        "  var b Owned<slice<uint8>>\n" +
        "  b = s.Bytes()\n" +
        "  e = s.Err()\n" +
        "  _ = ok; _ = txt; _ = b; _ = e\n" +
        "}\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    path := filepath.Join("build", "debug", "ir", "app", "u.ir.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    fns, _ := obj["functions"].([]any)
    if len(fns) == 0 { t.Fatalf("no functions in IR: %v", obj) }
    want := map[string]bool{
        "ami_rt_bufio_scanner_scan":  false,
        "ami_rt_bufio_scanner_text":  false,
        "ami_rt_bufio_scanner_bytes": false,
        "ami_rt_bufio_scanner_err":   false,
    }
    for _, f := range fns {
        fn := f.(map[string]any)
        blks := fn["blocks"].([]any)
        for _, bb := range blks {
            instrs := bb.(map[string]any)["instrs"].([]any)
            for _, in := range instrs {
                mo := in.(map[string]any)
                if mo["op"] != "EXPR" { continue }
                ex := mo["expr"].(map[string]any)
                if ex["op"] == "call" {
                    if cal, ok := ex["callee"].(string); ok {
                        if _, ok2 := want[cal]; ok2 { want[cal] = true }
                    }
                }
            }
        }
    }
    for k, v := range want {
        if !v { t.Fatalf("missing call callee %s in IR: %s", k, string(b)) }
    }
}
