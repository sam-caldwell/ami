package math

import (
    stdmath "math"
    "testing"
)

func approx(a, b, eps float64) bool { return stdmath.Abs(a-b) <= eps }

func TestConstants(t *testing.T) {
    if !approx(Pi, 3.1415926535, 1e-10) { t.Fatalf("Pi mismatch: %v", Pi) }
    if !approx(E, 2.7182818284, 1e-10) { t.Fatalf("E mismatch: %v", E) }
    if !approx(Phi, 1.6180339887, 1e-10) { t.Fatalf("Phi mismatch: %v", Phi) }
    if !approx(Sqrt2, 1.4142135623, 1e-10) { t.Fatalf("Sqrt2 mismatch: %v", Sqrt2) }
    if !approx(Ln2, 0.6931471805, 1e-10) { t.Fatalf("Ln2 mismatch: %v", Ln2) }
    if !approx(Ln10, 2.3025850929, 1e-10) { t.Fatalf("Ln10 mismatch: %v", Ln10) }
    if !approx(Log2E, 1.4426950409, 1e-10) { t.Fatalf("Log2E mismatch: %v", Log2E) }
    if !approx(Log10E, 0.4342944819, 1e-10) { t.Fatalf("Log10E mismatch: %v", Log10E) }
}

func TestAbsMaxMinAndNaN(t *testing.T) {
    if Abs(-5) != 5 { t.Fatalf("Abs(-5)") }
    if Max(1, 2) != 2 { t.Fatalf("Max") }
    if Min(1, 2) != 1 { t.Fatalf("Min") }
    if !IsNaN(Max(NaN(), 1)) { t.Fatalf("Max should propagate NaN") }
    if !IsNaN(Min(1, NaN())) { t.Fatalf("Min should propagate NaN") }
}

func TestRounding(t *testing.T) {
    if Ceil(1.2) != 2 { t.Fatalf("Ceil") }
    if Floor(1.8) != 1 { t.Fatalf("Floor") }
    if Trunc(-1.8) != -1 { t.Fatalf("Trunc") }
    if Round(2.5) != 3 { t.Fatalf("Round") }
    if RoundToEven(2.5) != 2 { t.Fatalf("RoundToEven ties-to-even") }
}

func TestModVsRemainder(t *testing.T) {
    if Mod(-3, 2) != -1 { t.Fatalf("Mod sign follows dividend") }
    if Remainder(-3, 2) != 1 { t.Fatalf("IEEE remainder differs at ties") }
}

func TestExponentialsAndLogs(t *testing.T) {
    if !approx(Exp(1), E, 1e-12) { t.Fatalf("Exp(1)") }
    if !approx(Expm1(1), E-1, 1e-12) { t.Fatalf("Expm1(1)") }
    if !approx(Exp2(3), 8, 1e-12) { t.Fatalf("Exp2(3)") }

    if !approx(Log(E), 1, 1e-12) { t.Fatalf("Log(E)") }
    if !approx(Log10(1000), 3, 1e-12) { t.Fatalf("Log10(1000)") }
    if !approx(Log1p(1), Log(2), 1e-12) { t.Fatalf("Log1p(1)") }
    if !approx(Log2(8), 3, 1e-12) { t.Fatalf("Log2(8)") }
}

func TestPowersAndRoots(t *testing.T) {
    if !approx(Pow(2, 10), 1024, 1e-12) { t.Fatalf("Pow") }
    if Pow10(3) != 1000 { t.Fatalf("Pow10") }
    if !approx(Sqrt(9), 3, 1e-12) { t.Fatalf("Sqrt") }
}

func TestTrigAndInverse(t *testing.T) {
    s, c := Sincos(Pi/3)
    if !approx(s, stdmath.Sqrt(3)/2, 1e-12) || !approx(c, 0.5, 1e-12) { t.Fatalf("Sincos") }
    if !approx(Sin(Pi/6), 0.5, 1e-12) { t.Fatalf("Sin") }
    if !approx(Cos(Pi/3), 0.5, 1e-12) { t.Fatalf("Cos") }
    if !approx(Tan(Pi/4), 1, 1e-12) { t.Fatalf("Tan") }

    if !approx(Asin(0.5), Pi/6, 1e-12) { t.Fatalf("Asin") }
    if !approx(Acos(0.5), Pi/3, 1e-12) { t.Fatalf("Acos") }
    if !approx(Atan(1), Pi/4, 1e-12) { t.Fatalf("Atan") }
    if !approx(Atan2(1, 1), Pi/4, 1e-12) { t.Fatalf("Atan2") }
}

func TestHyperbolic(t *testing.T) {
    if !approx(Tanh(0), 0, 1e-12) { t.Fatalf("Tanh(0)") }
    if !approx(Sinh(0), 0, 1e-12) { t.Fatalf("Sinh(0)") }
    if !approx(Cosh(0), 1, 1e-12) { t.Fatalf("Cosh(0)") }
}

func TestSpecialValuesAndClassification(t *testing.T) {
    nan := NaN()
    if !IsNaN(nan) { t.Fatalf("NaN classification") }
    posInf := Inf(+1)
    negInf := Inf(-1)
    if !IsInf(posInf, +1) || !IsInf(negInf, -1) { t.Fatalf("IsInf classification") }
    negZero := Copysign(0, -1)
    posZero := Copysign(0, +1)
    if !Signbit(negZero) || Signbit(posZero) { t.Fatalf("Signbit zero handling") }
    if Copysign(1, -2) != -1 { t.Fatalf("Copysign") }
    if Nextafter(1, 1) != 1 { t.Fatalf("Nextafter same") }
    if !(Nextafter(1, 2) > 1) { t.Fatalf("Nextafter toward +inf") }
}

func TestDecompositionHelpers(t *testing.T) {
    frac, exp := Frexp(8)
    // 8 == frac * 2^exp -> frac in [0.5, 1)
    if !(approx(frac, 0.5, 1e-12) && exp == 4) { t.Fatalf("Frexp") }
    if Ldexp(frac, exp) != 8 { t.Fatalf("Ldexp") }
    ip, fp := Modf(3.25)
    if ip != 3 || !approx(fp, 0.25, 1e-12) { t.Fatalf("Modf") }
}
