package driver

import (
    "strings"
    "testing"
)

func TestStdlibMath_Roundings_Intrinsics(t *testing.T) {
    s := buildAndReadLL(t, "mr1", "package app\nimport math\nfunc F() (float64){ var a float64 = math.Ceil(1.2); var b float64 = math.Floor(1.2); var c float64 = math.Trunc(1.2); var d float64 = math.Round(1.2); var e float64 = math.RoundToEven(2.5); return a+b+c+d+e }\n")
    for _, want := range []string{
        "call double @llvm.ceil.f64",
        "call double @llvm.floor.f64",
        "call double @llvm.trunc.f64",
        "call double @llvm.round.f64",
        "call double @llvm.roundeven.f64",
    } {
        if !strings.Contains(s, want) { t.Fatalf("missing rounding intrinsic %q:\n%s", want, s) }
    }
}

func TestStdlibMath_ExpLogPow_Intrinsics(t *testing.T) {
    s := buildAndReadLL(t, "mr2", "package app\nimport math\nfunc F() (float64){ var a float64 = math.Exp(1.0); var b float64 = math.Expm1(1.0); var c float64 = math.Exp2(2.0); var d float64 = math.Log(3.0); var e float64 = math.Log2(4.0); var f float64 = math.Log10(10.0); var g float64 = math.Pow(2.0, 8.0); return a+b+c+d+e+f+g }\n")
    for _, want := range []string{
        "call double @llvm.exp.f64",
        "call double @llvm.expm1.f64",
        "call double @llvm.exp2.f64",
        "call double @llvm.log.f64",
        "call double @llvm.log2.f64",
        "call double @llvm.log10.f64",
        "call double @llvm.pow.f64",
    } {
        if !strings.Contains(s, want) { t.Fatalf("missing exp/log/pow intrinsic %q:\n%s", want, s) }
    }
}

func TestStdlibMath_TrigAndInverse_Intrinsics(t *testing.T) {
    s := buildAndReadLL(t, "mr3", "package app\nimport math\nfunc F() (float64){ var a float64 = math.Sin(0.5); var b float64 = math.Cos(0.5); var c float64 = math.Tan(0.5); var d float64 = math.Asin(0.5); var e float64 = math.Acos(0.5); var f float64 = math.Atan(0.5); return a+b+c+d+e+f }\n")
    for _, want := range []string{
        "call double @llvm.sin.f64",
        "call double @llvm.cos.f64",
        "call double @llvm.tan.f64",
        "call double @llvm.asin.f64",
        "call double @llvm.acos.f64",
        "call double @llvm.atan.f64",
    } {
        if !strings.Contains(s, want) { t.Fatalf("missing trig/inv intrinsic %q:\n%s", want, s) }
    }
}

func TestStdlibMath_HyperbolicAndNext_Intrinsics(t *testing.T) {
    s := buildAndReadLL(t, "mr4", "package app\nimport math\nfunc F() (float64){ var a float64 = math.Sinh(1.0); var b float64 = math.Cosh(1.0); var c float64 = math.Tanh(1.0); var d float64 = math.Nextafter(1.0, 2.0); return a+b+c+d }\n")
    for _, want := range []string{
        "call double @llvm.sinh.f64",
        "call double @llvm.cosh.f64",
        "call double @llvm.tanh.f64",
        "call double @llvm.nextafter.f64",
    } {
        if !strings.Contains(s, want) { t.Fatalf("missing hyperbolic/nextafter intrinsic %q:\n%s", want, s) }
    }
}

