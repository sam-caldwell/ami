package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// builtinStdlibPackages returns AMI stdlib packages. It prefers loading from
// filesystem path `std/ami/stdlib/<pkg>/*.ami`. If none are found, it falls
// back to minimal in-memory stubs for critical packages.
func builtinStdlibPackages() []Package {
    if pkgs := fsStdlibPackages("std/ami/stdlib"); len(pkgs) > 0 {
        return pkgs
    }
    // Fallback: in-memory stubs to keep tests working without FS stdlib.
    var out []Package
    // time package stubs (signatures only).
    timeSrc := "package time\n" +
        "// AMI stdlib stubs (signatures only)\n" +
        "// Duration represents elapsed time (ns).\n" +
        "func Sleep(d Duration) {}\n" +
        "func Now() (Time) {}\n" +
        "func Add(t Time, d Duration) (Time) {}\n" +
        "func Delta(a Time, b Time) (int64) {}\n" +
        "func Unix(t Time) (int64) {}\n" +
        "func UnixNano(t Time) (int64) {}\n" +
        "func NewTicker(d Duration) (Ticker) {}\n" +
        "func TickerStart(t Ticker) {}\n" +
        "func TickerStop(t Ticker) {}\n" +
        "func TickerRegister(t Ticker, fn any) {}\n"
    tfs := &source.FileSet{}
    tfs.AddFile("time.ami", timeSrc)
    out = append(out, Package{Name: "time", Files: tfs})

    // signal package minimal surface
    sigSrc := "package signal\n" +
        "enum SignalType { SIGINT, SIGTERM, SIGHUP, SIGQUIT }\n" +
        "func Register(sig SignalType, fn any) {}\n" +
        "func Enable(sig SignalType) {}\n" +
        "func Disable(sig SignalType) {}\n" +
        "func Install(fn any) {}\n" +
        "func Token(fn any) (int64) {}\n"
    sfs := &source.FileSet{}
    sfs.AddFile("signal.ami", sigSrc)
    out = append(out, Package{Name: "signal", Files: sfs})

    // gpu package
    gpuSrc := "package gpu\n" +
        "func CudaAvailable() (bool) {}\n" +
        "func MetalAvailable() (bool) {}\n" +
        "func OpenCLAvailable() (bool) {}\n"
    gfs := &source.FileSet{}
    gfs.AddFile("gpu.ami", gpuSrc)
    out = append(out, Package{Name: "gpu", Files: gfs})

    // bufio package minimal stubs (signatures only)
    // TODO: When method receivers and cross-package type references are implemented,
    // replace these function-shaped APIs with true methods on Reader/Writer/Scanner.
    bufioSrc := "package bufio\n" +
        "// Minimal signatures to allow compiling AMI code importing bufio in tests.\n" +
        "type Reader struct{}\n" +
        "type Writer struct{}\n" +
        "type Scanner struct{}\n" +
        "func NewReader(src any) (Reader, error) {}\n" +
        // single-return variants to simplify lowering tests
        "func NewReaderSingle(src any) (Reader) {}\n" +
        // Method-style Reader APIs
        "func (r Reader) Read(n int) (Owned<slice<uint8>>, error) {}\n" +
        "func (r Reader) Peek(n int) (Owned<slice<uint8>>, error) {}\n" +
        "func (r Reader) UnreadByte() (error) {}\n" +
        // Function-style shims retained temporarily for older tests
        "func ReaderRead(r any, n int) (Owned<slice<uint8>>, error) {}\n" +
        "func ReaderPeek(r any, n int) (Owned<slice<uint8>>, error) {}\n" +
        "func ReaderUnreadByte(r any) (error) {}\n" +
        "func NewWriter(dst any) (Writer, error) {}\n" +
        // single-return variants to simplify lowering tests
        "func NewWriterSingle(dst any) (Writer) {}\n" +
        // Method-style Writer APIs
        "func (w Writer) Write(p Owned<slice<uint8>>) (int, error) {}\n" +
        "func (w Writer) Flush() (error) {}\n" +
        // Function-style shims retained temporarily
        "func WriterWrite(w any, p Owned<slice<uint8>>) (int, error) {}\n" +
        "func WriterFlush(w any) (error) {}\n" +
        "func NewScanner(r any) (Scanner, error) {}\n" +
        // Method-style Scanner APIs
        "func (s Scanner) Scan() (bool) {}\n" +
        "func (s Scanner) Text() (string) {}\n" +
        "func (s Scanner) Bytes() (Owned<slice<uint8>>) {}\n" +
        "func (s Scanner) Err() (error) {}\n" +
        // Function-style shims retained temporarily
        "func ScannerScan(s any) (bool) {}\n" +
        "func ScannerText(s any) (string) {}\n" +
        "func ScannerBytes(s any) (Owned<slice<uint8>>) {}\n" +
        "func ScannerErr(s any) (error) {}\n"
    bfs := &source.FileSet{}
    bfs.AddFile("bufio.ami", bufioSrc)
    out = append(out, Package{Name: "bufio", Files: bfs})

    return out
}

// fsStdlibPackages loads stdlib packages from the given root directory.
// It expects structure: <root>/<pkg>/*.ami. Returns packages sorted by name.
// fsStdlibPackages moved to stdlib_fs.go
