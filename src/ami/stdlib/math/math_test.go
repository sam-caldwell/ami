package math

import (
	stdmath "math"
	"testing"
)

func approxEqual(a, b, tol float64) bool {
	if stdmath.IsNaN(a) && stdmath.IsNaN(b) {
		return true
	}
	if stdmath.IsInf(a, 0) || stdmath.IsInf(b, 0) {
		return a == b
	}
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= tol
}

func TestMath_Constants(t *testing.T) {
	if !approxEqual(Pi, stdmath.Pi, 1e-15) {
		t.Fatal("Pi")
	}
	if !approxEqual(E, stdmath.E, 1e-15) {
		t.Fatal("E")
	}
	if !approxEqual(Sqrt2, stdmath.Sqrt2, 1e-15) {
		t.Fatal("Sqrt2")
	}
	if !approxEqual(Log2E, stdmath.Log2E, 1e-15) {
		t.Fatal("Log2E")
	}
	if !approxEqual(Log10E, stdmath.Log10E, 1e-15) {
		t.Fatal("Log10E")
	}
	if !approxEqual(Ln2, stdmath.Ln2, 1e-15) {
		t.Fatal("Ln2")
	}
	if !approxEqual(Ln10, stdmath.Ln10, 1e-15) {
		t.Fatal("Ln10")
	}
}

func TestMath_Functions_Happy(t *testing.T) {
	if Abs(-3.5) != 3.5 {
		t.Fatal("Abs")
	}
	if Min(2, 3) != 2 {
		t.Fatal("Min")
	}
	if Max(2, 3) != 3 {
		t.Fatal("Max")
	}
	if Floor(3.7) != 3 {
		t.Fatal("Floor")
	}
	if Ceil(3.2) != 4 {
		t.Fatal("Ceil")
	}
	if Trunc(-3.7) != -3 {
		t.Fatal("Trunc")
	}
	if Round(2.5) != 3 {
		t.Fatal("Round half away from zero")
	}
	if Round(-2.5) != -3 {
		t.Fatal("Round negative half away from zero")
	}
	if !approxEqual(Sqrt(9), 3, 1e-15) {
		t.Fatal("Sqrt")
	}
	if !approxEqual(Pow(2, 8), 256, 1e-15) {
		t.Fatal("Pow")
	}
}

func TestMath_EdgeCases_NaN_Inf(t *testing.T) {
	if !stdmath.IsNaN(Sqrt(-1)) {
		t.Fatal("Sqrt(-1) expected NaN")
	}
	if !stdmath.IsNaN(Min(stdmath.NaN(), 1)) {
		t.Fatal("Min with NaN")
	}
	if !stdmath.IsNaN(Max(stdmath.NaN(), 1)) {
		t.Fatal("Max with NaN")
	}
	if !stdmath.IsInf(Max(stdmath.Inf(1), 5), 1) {
		t.Fatal("Inf handling")
	}
}
