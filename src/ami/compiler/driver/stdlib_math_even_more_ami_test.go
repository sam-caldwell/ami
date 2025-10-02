package driver

import (
	"strings"
	"testing"
)

func testStdlibMath_FMA_Erf_Erfc_Intrinsics(t *testing.T) {
	s := buildAndReadLL(t, "mm1", "package app\nimport math\nfunc F() (float64){ var a float64 = math.FMA(1.0,2.0,3.0); var b float64 = math.Erf(1.0); var c float64 = math.Erfc(1.0); return a+b+c }\n")
	for _, want := range []string{
		"call double @llvm.fma.f64",
		"call double @llvm.erf.f64",
		"call double @llvm.erfc.f64",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing intrinsic %q:\n%s", want, s)
		}
	}
}

func testStdlibMath_Asinh_Acosh_Atanh_Runtime(t *testing.T) {
	s := buildAndReadLL(t, "mm2", "package app\nimport math\nfunc F() (float64){ var a float64 = math.Asinh(1.0); var b float64 = math.Acosh(2.0); var c float64 = math.Atanh(0.5); return a+b+c }\n")
	for _, want := range []string{
		"call double @ami_rt_math_asinh",
		"call double @ami_rt_math_acosh",
		"call double @ami_rt_math_atanh",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing runtime call %q:\n%s", want, s)
		}
	}
}

func testStdlibMath_Hypot_Cbrt_Dim_Runtime(t *testing.T) {
	s := buildAndReadLL(t, "mm3", "package app\nimport math\nfunc F() (float64){ var a float64 = math.Hypot(3.0,4.0); var b float64 = math.Cbrt(8.0); var c float64 = math.Dim(2.0,5.0); return a+b+c }\n")
	for _, want := range []string{
		"call double @ami_rt_math_hypot",
		"call double @ami_rt_math_cbrt",
		"call double @ami_rt_math_dim",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing runtime call %q:\n%s", want, s)
		}
	}
}

func testStdlibMath_Logb_Ilogb_Runtime(t *testing.T) {
	s := buildAndReadLL(t, "mm4", "package app\nimport math\nfunc F() (float64, int64){ var a float64 = math.Logb(8.0); var b int64 = math.Ilogb(8.0); return a, b }\n")
	if !strings.Contains(s, "call double @ami_rt_math_logb") {
		t.Fatalf("missing runtime call for logb:\n%s", s)
	}
	if !strings.Contains(s, "call i64 @ami_rt_math_ilogb") {
		t.Fatalf("missing runtime call for ilogb:\n%s", s)
	}
}
