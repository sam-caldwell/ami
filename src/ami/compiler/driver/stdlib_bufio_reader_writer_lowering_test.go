package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify bufio.Reader/Writer single-return constructors lower to calls and subsequent ops are present.
func TestStdlib_Bufio_ReaderWriter_Lowering_Callees(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\n" +
        "import bufio\n" +
        "func F(){\n" +
        "  var a any\n" +
        "  a = bufio.NewReaderSingle(a)\n" +
        "  bufio.ReaderPeek(a, 1)\n" +
        "  bufio.ReaderRead(a, 2)\n" +
        "  _ = bufio.ReaderUnreadByte(a)\n" +
        "  a = bufio.NewWriterSingle(a)\n" +
        "  var b Owned<slice<uint8>>\n" +
        "  bufio.WriterWrite(a, b)\n" +
        "  _ = bufio.WriterFlush(a)\n" +
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
        "bufio.NewReaderSingle": false,
        "bufio.ReaderPeek":     false,
        "bufio.ReaderRead":     false,
        "bufio.ReaderUnreadByte": false,
        "bufio.NewWriterSingle": false,
        "bufio.WriterWrite":    false,
        "bufio.WriterFlush":    false,
    }
    for _, f := range fns {
        fn := f.(map[string]any)
        blks := fn["blocks"].([]any)
        for _, bb := range blks {
            instrs := bb.(map[string]any)["instrs"].([]any)
            for _, in := range instrs {
                mo := in.(map[string]any)
                if mo["op"].(string) != "EXPR" { continue }
                ex := mo["expr"].(map[string]any)
                if ex["op"].(string) != "call" { continue }
                if cal, ok := ex["callee"].(string); ok {
                    if _, ok2 := want[cal]; ok2 { want[cal] = true }
                }
            }
        }
    }
    for k, v := range want {
        if !v { t.Fatalf("missing call callee %s in IR: %s", k, string(b)) }
    }
}
