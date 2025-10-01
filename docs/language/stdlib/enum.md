# Stdlib: enum (Descriptor‑Driven Enums)

The `enum` module provides small, descriptor‑driven helpers for generated enums, with deterministic JSON/text behavior.

API (AMI module `enum`)
- `type Descriptor { name string, names slice<string> }` — enum type name and canonical member names in ordinal order.
- `func enum.descriptor(typeName string, names slice<string>) (Descriptor, error)` — construct and validate a descriptor.
- `func enum.mustDescriptor(typeName string, names slice<string>) Descriptor` — panic on invalid input (for generated code).
- `func enum.string(d Descriptor, v int) string` — canonical name for ordinal `v` (empty if invalid).
- `func enum.ordinal(v int) int` — identity helper; returns `v`.
- `func enum.isValid(d Descriptor, v int) bool` — true when ordinal is within range.
- `func enum.values(d Descriptor) slice<int>` — all ordinals `[0..N-1]`.
- `func enum.names(d Descriptor) slice<string>` — copy of canonical names.
- `func enum.parse(d Descriptor, s string) (int, error)` — name → ordinal (case‑sensitive).
- `func enum.mustParse(d Descriptor, s string) int` — panic if `s` invalid.
- `func enum.fromOrdinal(d Descriptor, i int) (int, error)` — validate ordinal and return it.

Diagnostics
- Errors use deterministic messages for parsing/ordinal failures (e.g., `E_ENUM_PARSE`, `E_ENUM_ORDINAL`).

Examples (AMI)
```
var d = enum.descriptor("Color", ["Red", "Green", "Blue"])
var g = enum.mustParse(d, "Green") // 1
var ok = enum.isValid(d, g)         // true
var name = enum.string(d, g)        // "Green"
```
