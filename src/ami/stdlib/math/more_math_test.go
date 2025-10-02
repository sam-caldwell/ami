package math

import (
    stdmath "math"
    "testing"
)

func TestExtras_FMA_Erf_Erfc_Hypot_Cbrt_Asinh_Acosh_Atanh_Dim_Logb_Ilogb(t *testing.T) {
    if FMA(2, 3, 4) != 10 { t.Fatalf("FMA") }
    if !approx(Erf(0.5), 0.5204998778, 1e-9) { t.Fatalf("Erf") }
    if !approx(Erfc(0.5), 0.4795001221, 1e-9) { t.Fatalf("Erfc") }
    if !approx(Hypot(3, 4), 5, 1e-12) { t.Fatalf("Hypot") }
    if !approx(Cbrt(8), 2, 1e-12) { t.Fatalf("Cbrt") }
    if !approx(Asinh(1), stdmath.Asinh(1), 1e-12) { t.Fatalf("Asinh") }
    if !approx(Acosh(2), stdmath.Acosh(2), 1e-12) { t.Fatalf("Acosh") }
    if !approx(Atanh(0.5), stdmath.Atanh(0.5), 1e-12) { t.Fatalf("Atanh") }
    if !approx(Dim(5, 3), 2, 1e-12) { t.Fatalf("Dim") }
    if !approx(Logb(8), stdmath.Logb(8), 1e-12) { t.Fatalf("Logb") }
    if Ilogb(8) != stdmath.Ilogb(8) { t.Fatalf("Ilogb") }
}

