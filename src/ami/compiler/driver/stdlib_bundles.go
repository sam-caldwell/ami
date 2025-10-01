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

    // gpu package: top-level availability probes only (stubs). Additional APIs provided by Go stdlib.
    gpuSrc := "package gpu\n" +
        "// AMI stdlib stubs (signatures only)\n" +
        "func CudaAvailable() (bool) {}\n" +
        "func MetalAvailable() (bool) {}\n" +
        "func OpenCLAvailable() (bool) {}\n" +
        "// BlockingSubmit wraps GPU submission and blocks until completion, returning an Error<any>\n" +
        "func BlockingSubmit(x any) (Error<any>) {}\n"
    gfs := &source.FileSet{}
    gfs.AddFile("gpu.ami", gpuSrc)
    out = append(out, Package{Name: "gpu", Files: gfs})

    // math package: surface of float math operations and helpers (stubs only).
    // Use 'float64' for floating point type in AMI; booleans/ints for control parameters.
    mathSrc := "package math\n" +
        "// AMI stdlib stubs (signatures only)\n" +
        // Basics
        "func Abs(x float64) (float64) {}\n" +
        "func Max(x float64, y float64) (float64) {}\n" +
        "func Min(x float64, y float64) (float64) {}\n" +
        // Rounding
        "func Ceil(x float64) (float64) {}\n" +
        "func Floor(x float64) (float64) {}\n" +
        "func Trunc(x float64) (float64) {}\n" +
        "func Round(x float64) (float64) {}\n" +
        "func RoundToEven(x float64) (float64) {}\n" +
        // Modulo and IEEE remainder
        "func Mod(x float64, y float64) (float64) {}\n" +
        "func Remainder(x float64, y float64) (float64) {}\n" +
        // Exponentials and logs
        "func Exp(x float64) (float64) {}\n" +
        "func Expm1(x float64) (float64) {}\n" +
        "func Exp2(x float64) (float64) {}\n" +
        "func Log(x float64) (float64) {}\n" +
        "func Log10(x float64) (float64) {}\n" +
        "func Log1p(x float64) (float64) {}\n" +
        "func Log2(x float64) (float64) {}\n" +
        // Powers and roots
        "func Pow(x float64, y float64) (float64) {}\n" +
        "func Pow10(n int) (float64) {}\n" +
        "func Sqrt(x float64) (float64) {}\n" +
        // Trigonometry
        "func Sin(x float64) (float64) {}\n" +
        "func Cos(x float64) (float64) {}\n" +
        "func Tan(x float64) (float64) {}\n" +
        "func Sincos(x float64) (float64, float64) {}\n" +
        // Inverse trig
        "func Asin(x float64) (float64) {}\n" +
        "func Acos(x float64) (float64) {}\n" +
        "func Atan(x float64) (float64) {}\n" +
        "func Atan2(y float64, x float64) (float64) {}\n" +
        // Hyperbolic
        "func Sinh(x float64) (float64) {}\n" +
        "func Cosh(x float64) (float64) {}\n" +
        "func Tanh(x float64) (float64) {}\n" +
        // Special values and classification
        "func NaN() (float64) {}\n" +
        "func Inf(sign int) (float64) {}\n" +
        "func IsNaN(x float64) (bool) {}\n" +
        "func IsInf(x float64, sign int) (bool) {}\n" +
        "func Signbit(x float64) (bool) {}\n" +
        // Bitwise float helpers
        "func Copysign(x float64, y float64) (float64) {}\n" +
        "func Nextafter(x float64, y float64) (float64) {}\n" +
        // Decomposition
        "func Frexp(x float64) (float64, int) {}\n" +
        "func Ldexp(frac float64, exp int) (float64) {}\n" +
        "func Modf(x float64) (float64, float64) {}\n"
    mfs := &source.FileSet{}
    mfs.AddFile("math.ami", mathSrc)
    out = append(out, Package{Name: "math", Files: mfs})
    return out
}
