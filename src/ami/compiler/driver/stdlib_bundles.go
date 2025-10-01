package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// builtinStdlibPackages returns a slice of builtin AMI stdlib packages provided as in-memory stubs.
// These enable AMI users to import packages like `time` and `signal` without providing implementations.
func builtinStdlibPackages() []Package {
    var out []Package
    // time package stubs (signatures only). Units: int parameters represent duration in milliseconds for Sleep/Add.
    timeSrc := "package time\n" +
        "// AMI stdlib stubs (signatures only)\n" +
        "// Duration values are integers (ms) for the purpose of stubs.\n" +
        "func Sleep(d int) {}\n" +
        "func Now() (Time) {}\n" +
        "func Add(t Time, d int) (Time) {}\n" +
        "func Delta(a Time, b Time) (int64) {}\n" +
        "func Unix(t Time) (int64) {}\n" +
        "func UnixNano(t Time) (int64) {}\n"
    tfs := &source.FileSet{}
    tfs.AddFile("time.ami", timeSrc)
    out = append(out, Package{Name: "time", Files: tfs})

    // signal package minimal surface (SignalType and Register). Handlers are stubs; semantics handled by runtime.
    sigSrc := "package signal\n" +
        "enum SignalType { SIGINT, SIGTERM, SIGHUP, SIGQUIT }\n" +
        "func Register(sig SignalType, fn func()) {}\n"
    sfs := &source.FileSet{}
    sfs.AddFile("signal.ami", sigSrc)
    out = append(out, Package{Name: "signal", Files: sfs})
    return out
}

