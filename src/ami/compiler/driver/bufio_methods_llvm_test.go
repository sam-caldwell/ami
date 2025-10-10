package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify bufio method calls map to runtime shims in LLVM when rewrite is enabled.
func TestLower_Bufio_Methods_Emit_Runtime_Shim_Externs_And_Calls_InLLVM(t *testing.T) {
    // Enable mapping
    old := enableBufioMethodRewrite
    enableBufioMethodRewrite = true
    defer func() { enableBufioMethodRewrite = old }()

    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\n" +
        "import bufio\n" +
        "func F(){\n" +
        "  var src any\n" +
        "  var r bufio.Reader\n" +
        "  r = bufio.NewReaderSingle(src)\n" +
        "  r.Peek(1)\n" +
        "  r.Read(2)\n" +
        "  _ = r.UnreadByte()\n" +
        "  var w bufio.Writer\n" +
        "  w = bufio.NewWriterSingle(src)\n" +
        "  var b Owned<slice<uint8>>\n" +
        "  w.Write(b)\n" +
        "  _ = w.Flush()\n" +
        "}\n"
    fs.AddFile("bm.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "bm.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    // Ensure externs are present
    decls := []string{
        "declare { i64, i64 } @ami_rt_bufio_reader_read(i64, i64)",
        "declare { i64, i64 } @ami_rt_bufio_reader_peek(i64, i64)",
        "declare ptr @ami_rt_bufio_reader_unread_byte(i64)",
        "declare { i64, i64 } @ami_rt_bufio_writer_write(i64, i64)",
        "declare ptr @ami_rt_bufio_writer_flush(i64)",
    }
    for _, d := range decls { if !strings.Contains(s, d) { t.Fatalf("missing extern: %s\n%s", d, s) } }
    // Ensure calls are present
    calls := []string{
        "call { i64, i64 } @ami_rt_bufio_reader_peek(",
        "call { i64, i64 } @ami_rt_bufio_reader_read(",
        "call ptr @ami_rt_bufio_reader_unread_byte(",
        "call { i64, i64 } @ami_rt_bufio_writer_write(",
        "call ptr @ami_rt_bufio_writer_flush(",
    }
    for _, c := range calls { if !strings.Contains(s, c) { t.Fatalf("missing call: %s\n%s", c, s) } }
}

// Verify scanner method calls map to runtime shims in LLVM when rewrite is enabled.
func TestLower_Bufio_Scanner_Methods_Emit_Runtime_Shim_Externs_And_Calls_InLLVM(t *testing.T) {
    old := enableBufioMethodRewrite
    enableBufioMethodRewrite = true
    defer func() { enableBufioMethodRewrite = old }()

    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\n" +
        "import bufio\n" +
        "func F(){\n" +
        "  var s bufio.Scanner\n" +
        "  s.Scan()\n" +
        "  s.Text()\n" +
        "  s.Bytes()\n" +
        "  s.Err()\n" +
        "}\n"
    fs.AddFile("bs.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "bs.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    decls := []string{
        "declare i1 @ami_rt_bufio_scanner_scan(i64)",
        "declare ptr @ami_rt_bufio_scanner_text(i64)",
        "declare ptr @ami_rt_bufio_scanner_bytes(i64)",
        "declare ptr @ami_rt_bufio_scanner_err(i64)",
    }
    for _, d := range decls { if !strings.Contains(s, d) { t.Fatalf("missing extern: %s\n%s", d, s) } }
    calls := []string{
        "call i1 @ami_rt_bufio_scanner_scan(",
        "call ptr @ami_rt_bufio_scanner_text(",
        "call ptr @ami_rt_bufio_scanner_bytes(",
        "call ptr @ami_rt_bufio_scanner_err(",
    }
    for _, c := range calls { if !strings.Contains(s, c) { t.Fatalf("missing call: %s\n%s", c, s) } }
}
