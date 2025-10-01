package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// builtinStdlibPackages returns a slice of builtin AMI stdlib packages provided as in-memory stubs.
// These enable AMI users to import packages like `time` and `signal` without providing implementations.
func builtinStdlibPackages() []Package {
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
        // Ticker API surface (package-level functions; parser lacks method decls).\n" +
        "func NewTicker(d Duration) (Ticker) {}\n" +
        "func TickerStart(t Ticker) {}\n" +
        "func TickerStop(t Ticker) {}\n" +
        "func TickerRegister(t Ticker, fn any) {}\n"
    tfs := &source.FileSet{}
    tfs.AddFile("time.ami", timeSrc)
    out = append(out, Package{Name: "time", Files: tfs})

    // signal package minimal surface (SignalType and Register). Handlers are stubs; semantics handled by runtime.
    sigSrc := "package signal\n" +
        "enum SignalType { SIGINT, SIGTERM, SIGHUP, SIGQUIT }\n" +
        "// Use 'any' for handler to avoid function-typed params in parser scaffold\n" +
        "func Register(sig SignalType, fn any) {}\n" +
        "func Enable(sig SignalType) {}\n" +
        "func Disable(sig SignalType) {}\n" +
        "// Future handler primitives:\n" +
        "func Install(fn any) {}\n" +
        "func Token(fn any) (int64) {}\n"
    sfs := &source.FileSet{}
    sfs.AddFile("signal.ami", sigSrc)
    out = append(out, Package{Name: "signal", Files: sfs})

    // gpu package: top-level availability probes (signatures only)
    gpuSrc := "package gpu\n" +
        "// AMI stdlib stubs (signatures only)\n" +
        "func CudaAvailable() (bool) {}\n" +
        "func MetalAvailable() (bool) {}\n" +
        "func OpenCLAvailable() (bool) {}\n"
    gfs := &source.FileSet{}
    gfs.AddFile("gpu.ami", gpuSrc)
    out = append(out, Package{Name: "gpu", Files: gfs})

    return out
}
