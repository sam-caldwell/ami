package math

import (
    stdmath "math"
)

// Constants aligned with Go's math package.
const (
    E      = stdmath.E
    Pi     = stdmath.Pi
    Phi    = stdmath.Phi
    Sqrt2  = stdmath.Sqrt2
    Ln2    = stdmath.Ln2
    Ln10   = stdmath.Ln10
    Log2E  = stdmath.Log2E
    Log10E = stdmath.Log10E
)

// Basic operations
func Abs(x float64) float64             { return stdmath.Abs(x) }
func Max(x, y float64) float64          { return stdmath.Max(x, y) }
func Min(x, y float64) float64          { return stdmath.Min(x, y) }

// Rounding helpers
func Ceil(x float64) float64            { return stdmath.Ceil(x) }
func Floor(x float64) float64           { return stdmath.Floor(x) }
func Trunc(x float64) float64           { return stdmath.Trunc(x) }
func Round(x float64) float64           { return stdmath.Round(x) }
func RoundToEven(x float64) float64     { return stdmath.RoundToEven(x) }

// Modulo and remainder
func Mod(x, y float64) float64          { return stdmath.Mod(x, y) }
func Remainder(x, y float64) float64    { return stdmath.Remainder(x, y) }

// Exponentials
func Exp(x float64) float64             { return stdmath.Exp(x) }
func Expm1(x float64) float64           { return stdmath.Expm1(x) }
func Exp2(x float64) float64            { return stdmath.Exp2(x) }

// Logarithms
func Log(x float64) float64             { return stdmath.Log(x) }
func Log10(x float64) float64           { return stdmath.Log10(x) }
func Log1p(x float64) float64           { return stdmath.Log1p(x) }
func Log2(x float64) float64            { return stdmath.Log2(x) }

// Powers and roots
func Pow(x, y float64) float64          { return stdmath.Pow(x, y) }
func Pow10(n int) float64               { return stdmath.Pow10(n) }
func Sqrt(x float64) float64            { return stdmath.Sqrt(x) }

// Trigonometry
func Sin(x float64) float64             { return stdmath.Sin(x) }
func Cos(x float64) float64             { return stdmath.Cos(x) }
func Tan(x float64) float64             { return stdmath.Tan(x) }
func Sincos(x float64) (sin, cos float64) { return stdmath.Sincos(x) }

// Inverse trigonometry
func Asin(x float64) float64            { return stdmath.Asin(x) }
func Acos(x float64) float64            { return stdmath.Acos(x) }
func Atan(x float64) float64            { return stdmath.Atan(x) }
func Atan2(y, x float64) float64        { return stdmath.Atan2(y, x) }

// Hyperbolic
func Sinh(x float64) float64            { return stdmath.Sinh(x) }
func Cosh(x float64) float64            { return stdmath.Cosh(x) }
func Tanh(x float64) float64            { return stdmath.Tanh(x) }

// Special values
func NaN() float64                      { return stdmath.NaN() }
func Inf(sign int) float64              { return stdmath.Inf(sign) }

// Classification and manipulation
func IsNaN(x float64) bool              { return stdmath.IsNaN(x) }
func IsInf(x float64, sign int) bool    { return stdmath.IsInf(x, sign) }
func Signbit(x float64) bool            { return stdmath.Signbit(x) }
func Copysign(x, y float64) float64     { return stdmath.Copysign(x, y) }
func Nextafter(x, y float64) float64    { return stdmath.Nextafter(x, y) }

// Decomposition helpers
func Frexp(x float64) (frac float64, exp int) { return stdmath.Frexp(x) }
func Ldexp(frac float64, exp int) float64     { return stdmath.Ldexp(frac, exp) }
func Modf(x float64) (intPart, fracPart float64) { return stdmath.Modf(x) }

