# Stdlib: enum (Descriptor‑Driven Enums)

The `enum` stdlib package provides a small, descriptor‑driven model for enums with helpers for validation and JSON/Text encoding.

API (Go package `enum`)
- `type Descriptor struct { Name string; Names []string }`: describes the enum type and canonical member names (ordinal order).
- `NewDescriptor(typeName string, names []string) (Descriptor, error)`: validate and construct a descriptor.
- `MustNewDescriptor(typeName string, names []string) Descriptor`: panic on invalid input (for generated code).
- `String(d Descriptor, v int) string`: canonical name for ordinal `v` (empty if invalid).
- `Ordinal(v int) int`: identity helper; returns `v`.
- `IsValid(d Descriptor, v int) bool`: true when ordinal is within range.
- `Values(d Descriptor) []int`: all ordinals `[0..N-1]`.
- `Names(d Descriptor) []string`: copy of canonical names.
- `Parse(d Descriptor, s string) (int, error)`: name → ordinal (case‑sensitive).
- `MustParse(d Descriptor, s string) int`: panic if `s` invalid.
- `FromOrdinal(d Descriptor, i int) (int, error)`: validate ordinal and return it.
- `type Value struct { D Descriptor; V int }`: wrapper providing `String()`, `GoString()`, `MarshalJSON`, `UnmarshalJSON`, `MarshalText`, `UnmarshalText`.

Diagnostics
- Errors use deterministic messages for parsing/ordinal failures (e.g., `E_ENUM_PARSE`, `E_ENUM_ORDINAL`).

Examples
```go
d := enum.MustNewDescriptor("Color", []string{"Red", "Green", "Blue"})
g := enum.MustParse(d, "Green") // 1
ok := enum.IsValid(d, g)          // true
name := enum.String(d, g)         // "Green"

vv := enum.Value{D: d, V: g}
b, _ := vv.MarshalJSON()          // => "\"Green\""
_ = b
```

