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

    return out
}

// fsStdlibPackages loads stdlib packages from the given root directory.
// It expects structure: <root>/<pkg>/*.ami. Returns packages sorted by name.
// fsStdlibPackages moved to stdlib_fs.go
