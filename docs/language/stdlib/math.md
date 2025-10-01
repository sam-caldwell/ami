# math

The `math` stdlib package provides AMI’s language‑level mathematical functions and constants. The AMI runtime implements these in Go and mirrors Go's `math` semantics to ensure deterministic behavior and IEEE‑754 alignment. All functions operate on `float64` unless otherwise noted.

- Basics: `Abs(x)`, `Max(x, y)`, `Min(x, y)`
- Rounding: `Ceil(x)`, `Floor(x)`, `Trunc(x)`, `Round(x)`, `RoundToEven(x)`
- Modulo: `Mod(x, y)`, `Remainder(x, y)` (IEEE‑754 remainder)
- Exponentials: `Exp(x)`, `Expm1(x)`, `Exp2(x)`
- Logs: `Log(x)`, `Log10(x)`, `Log1p(x)`, `Log2(x)`
- Powers: `Pow(x, y)`, `Pow10(n int)`, `Sqrt(x)`
- Trig: `Sin(x)`, `Cos(x)`, `Tan(x)`, `Sincos(x)`
- Inverse Trig: `Asin(x)`, `Acos(x)`, `Atan(x)`, `Atan2(y, x)`
- Hyperbolic: `Sinh(x)`, `Cosh(x)`, `Tanh(x)`
- Special values: `NaN()`, `Inf(sign)`; classification: `IsNaN(x)`, `IsInf(x, sign)`, `Signbit(x)`
- Bitwise float helpers: `Copysign(x, y)`, `Nextafter(x, y)`
- Decomposition: `Frexp(x) (frac, exp)`, `Ldexp(frac, exp)`, `Modf(x) (ip, fp)`

Constants match Go's definitions: `Pi`, `E`, `Phi`, `Sqrt2`, `Ln2`, `Ln10`, `Log2E`, `Log10E`.

Usage (AMI)

- Import as `math` in AMI source; call like `math.Abs(x)`.
- The underlying runtime is opaque to AMI users; only AMI’s `math` surface is considered stable.

Semantics

- NaN/Inf: Functions propagate `NaN`/`Inf` as in Go's `math`. `Inf(+1)` is `+Inf`, `Inf(-1)` is `-Inf`.
- Remainder vs Mod: `Mod(x, y)` has the sign of `x`; `Remainder(x, y)` implements IEEE‑754 remainder (`x - n*y` where `n` is the nearest integer to `x/y`, ties to even).
- Determinism: All operations are thin wrappers over Go's `math` package and follow its corner‑case behavior.
