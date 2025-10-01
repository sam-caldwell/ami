# math

The `math` stdlib package provides AMI’s language‑level mathematical functions and constants. AMI lowers `math` calls directly to LLVM intrinsics or portable runtime helpers (emitted as LLVM IR) to ensure deterministic behavior and IEEE‑754 alignment across platforms. All functions operate on `float64` unless otherwise noted.

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
- Calls lower to LLVM: single‑result map to intrinsics (e.g., `llvm.sqrt.f64`, `llvm.maxnum.f64`); multi‑result
  (`Sincos`, `Frexp`, `Modf`) lower to runtime helpers returning aggregates and are consumed via `extractvalue`.

Semantics

- NaN/Inf: Functions propagate `NaN`/`Inf` consistent with Go's `math`. `Inf(+1)` is `+Inf`, `Inf(-1)` is `-Inf`.
- Remainder vs Mod: `Mod(x, y)` has the sign of `x` (emitted as `frem`); `Remainder(x, y)` uses a portable runtime helper implementing IEEE‑754 semantics.
- Determinism: Lowering is via LLVM intrinsics where available and portable helper IR otherwise; corner cases follow Go's `math`.
