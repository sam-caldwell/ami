# AMI Language: Enum Semantics

This document defines enums in AMI consistent with the authoritative reference (docs/Asynchronous Machine Interface.docx). It describes naming, determinism, methods, JSON/text behavior, and unknown handling. Compiler/runtime behavior must remain deterministic and case‑sensitive.

Overview

- Enum members have stable, case‑sensitive canonical names and zero‑based ordinals.
- Unknown members are not allowed in source; tools may expose a reserved sentinel for wiring tests only.
- Generated tables (names and values) use stable ordering; public outputs never rely on map iteration order.

Naming & Determinism

- Canonical names: exact member identifiers as written in source; case‑sensitive.
- Ordinals: assigned by declaration order starting at 0; stable across builds.
- Values()/Names(): return slices ordered by ordinal (lexical declaration order).

Methods (runtime/helpers or codegen)

- String() string: returns the canonical name; unknown values render as "<invalid>".
- Ordinal() int: returns zero‑based ordinal of the value.
- IsValid() bool: returns true when the value is within the known ordinal range and not an unknown sentinel.
- Values() []<Enum>: returns all enum values in canonical order.
- Names() []string: returns all canonical names in canonical order.
- Parse<Enum>(s string) (<Enum>, error): parses by canonical name. Case‑sensitive; returns E_ENUM_PARSE on invalid.
- MustParse<Enum>(s string) <Enum>: same as Parse but panics on invalid; suitable for tests/tooling.
- FromOrdinal(i int) (<Enum>, error): maps ordinal to value or returns E_ENUM_ORDINAL on invalid.
- JSON: MarshalJSON()/UnmarshalJSON() use canonical string names; invalid input returns E_ENUM_PARSE.
- Text: MarshalText()/UnmarshalText() use canonical names; invalid input returns E_ENUM_PARSE.
- GoString() string: returns stable debug form Enum(Name).

Unknown Handling

- Parsing invalid strings or ordinals returns well‑defined errors: E_ENUM_PARSE and E_ENUM_ORDINAL respectively.
- IsValid() guards: callers must check before using values; generated code avoids panics by design.

Examples

```
enum Color { Red, Green, Blue }

// String & ordinal
ColorRed.String()  // "Red"
ColorRed.Ordinal() // 0

// Parse
v, err := ParseColor("Green") // v=ColorGreen, err=nil
_, err = ParseColor("green")  // err=E_ENUM_PARSE (case‑sensitive)

// JSON round‑trip
b, _ := json.Marshal(ColorBlue) // "\"Blue\""
var x Color
_ = json.Unmarshal(b, &x)       // x=ColorBlue

// FromOrdinal
x, _ = FromOrdinal(2)           // x=ColorBlue
_, err = FromOrdinal(3)         // err=E_ENUM_ORDINAL

// Names/Values are canonical order
Names()  // ["Red","Green","Blue"]
Values() // [ColorRed, ColorGreen, ColorBlue]
```

Guarantees

- Deterministic: Names/Values ordering and String/Ordinal mappings are stable given identical source.
- Case‑sensitive: Parse rejects mismatched case.
- No map iteration: public outputs sourced from ordered slices, not maps.

Testing Guidance

- String/JSON/Text round‑trips for all members.
- Parse and FromOrdinal errors on invalid inputs.
- Names()/Values() ordering matches declaration order.
- IsValid() true for known values; false for out of range and reserved unknown sentinel (if present).

