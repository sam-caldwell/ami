package math

import stdmath "math"

// Abs returns the absolute value of x.
func Abs(x float64) float64 { return stdmath.Abs(x) }

// Min returns the smaller of x or y.
func Min(x, y float64) float64 { return stdmath.Min(x, y) }

// Max returns the larger of x or y.
func Max(x, y float64) float64 { return stdmath.Max(x, y) }

// Floor returns the greatest integer value less than or equal to x.
func Floor(x float64) float64 { return stdmath.Floor(x) }

// Ceil returns the least integer value greater than or equal to x.
func Ceil(x float64) float64 { return stdmath.Ceil(x) }

// Trunc returns the integer value of x truncated toward zero.
func Trunc(x float64) float64 { return stdmath.Trunc(x) }

// Round returns the nearest integer, rounding half away from zero.
func Round(x float64) float64 { return stdmath.Round(x) }

// Pow returns x**y, the base-x exponential of y.
func Pow(x, y float64) float64 { return stdmath.Pow(x, y) }

// Sqrt returns the square root of x.
func Sqrt(x float64) float64 { return stdmath.Sqrt(x) }

