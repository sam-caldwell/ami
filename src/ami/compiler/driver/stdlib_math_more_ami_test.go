package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func buildAndReadLL(t *testing.T, unit, src string) string {
    t.Helper()
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile(unit+".ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", unit+".ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    return string(b)
}

func TestStdlibMath_Log1p_Atan2_Cast(t *testing.T) {
    s := buildAndReadLL(t, "m1", "package app\nimport math\nfunc F() (float64){ var a float64 = math.Atan2(2.0,1.0); return math.Log1p(a) }\n")
    if !strings.Contains(s, "call double @llvm.atan2.f64") { t.Fatalf("missing atan2 intrinsic:\n%s", s) }
    if !strings.Contains(s, "call double @llvm.log1p.f64") { t.Fatalf("missing log1p intrinsic:\n%s", s) }
}

func TestStdlibMath_Copysign_Nextafter_Ldexp(t *testing.T) {
    s := buildAndReadLL(t, "m2", "package app\nimport math\nfunc F() (float64){ var x float64 = math.Copysign(1.0, -2.0); return math.Ldexp(x, 3) }\n")
    if !strings.Contains(s, "call double @llvm.copysign.f64") { t.Fatalf("missing copysign intrinsic:\n%s", s) }
    if !strings.Contains(s, "call double @llvm.ldexp.f64") { t.Fatalf("missing ldexp intrinsic:\n%s", s) }
}

func TestStdlibMath_Mod_Remainder(t *testing.T) {
    s1 := buildAndReadLL(t, "m3", "package app\nimport math\nfunc F() (float64){ return math.Mod(5.0, 2.0) }\n")
    if !strings.Contains(s1, " = frem double ") { t.Fatalf("missing frem for math.Mod:\n%s", s1) }
    s2 := buildAndReadLL(t, "m4", "package app\nimport math\nfunc F() (float64){ return math.Remainder(5.0, 2.0) }\n")
    if !strings.Contains(s2, "call double @ami_rt_math_remainder") { t.Fatalf("missing runtime remainder call:\n%s", s2) }
}

func TestStdlibMath_NaN_Inf_IsNaN_IsInf_Signbit(t *testing.T) {
    s := buildAndReadLL(t, "m5", "package app\nimport math\nfunc F() (float64){ var x float64 = 1.0; var y float64 = 2.0; if math.IsNaN(math.NaN()) { return math.Inf(1) } else { if math.IsInf(x, 1) { return y } else { if math.Signbit(-1.0) { return y } else { return x } } } }\n")
    if !strings.Contains(s, "call double @ami_rt_math_nan") { t.Fatalf("missing math_nan runtime:\n%s", s) }
    if !strings.Contains(s, "call double @ami_rt_math_inf") { t.Fatalf("missing math_inf runtime:\n%s", s) }
    if !strings.Contains(s, "call i1 @ami_rt_math_isnan") { t.Fatalf("missing isnan runtime:\n%s", s) }
    if !strings.Contains(s, "call i1 @ami_rt_math_isinf") { t.Fatalf("missing isinf runtime:\n%s", s) }
    if !strings.Contains(s, "call i1 @ami_rt_math_signbit") { t.Fatalf("missing signbit runtime:\n%s", s) }
}

func TestStdlibMath_Pow10_Sincos_Frexp_Modf(t *testing.T) {
    s1 := buildAndReadLL(t, "m6", "package app\nimport math\nfunc F() (float64){ return math.Pow10(3) }\n")
    if !strings.Contains(s1, "call double @ami_rt_math_pow10") { t.Fatalf("missing pow10 runtime:\n%s", s1) }
    s2 := buildAndReadLL(t, "m7", "package app\nimport math\nfunc F() (float64, float64){ return math.Sincos(1.0) }\n")
    if !strings.Contains(s2, "call { double, double } @ami_rt_math_sincos") { t.Fatalf("missing sincos runtime aggregate:\n%s", s2) }
    if !strings.Contains(s2, "extractvalue { double, double }") { t.Fatalf("missing extractvalue for sincos:\n%s", s2) }
    s3 := buildAndReadLL(t, "m8", "package app\nimport math\nfunc F() (float64, int64){ return math.Frexp(8.0) }\n")
    if !strings.Contains(s3, "call { double, i64 } @ami_rt_math_frexp") { t.Fatalf("missing frexp runtime aggregate:\n%s", s3) }
    s4 := buildAndReadLL(t, "m9", "package app\nimport math\nfunc F() (float64, float64){ return math.Modf(3.14) }\n")
    if !strings.Contains(s4, "call { double, double } @ami_rt_math_modf") { t.Fatalf("missing modf runtime aggregate:\n%s", s4) }
}
