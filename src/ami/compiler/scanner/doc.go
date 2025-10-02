// Package scanner converts AMI source text into a stream of tokens.
//
// Responsibilities
//
//   - Unicode-aware lexical analysis with deterministic tokenization.
//   - Classification of keywords, operators, and literals (identifiers,
//     numbers, strings, durations, etc.).
//   - Position tracking integrated with the source package for precise
//     diagnostics and recovery ranges.
//   - Skips whitespace and comments while preserving any newline significance
//     required by the grammar.
//
// Inputs/Outputs
//
//   - Input: in-memory source buffers (no implicit I/O).
//   - Output: a token stream consumed by the parser and early diagnostics for
//     malformed lexemes; errors are recoverable to enable best-effort parsing.
//
// Invariants
//
//   - Deterministic: identical input produces identical token streams and
//     diagnostic ordering.
//   - Side-effect free: no filesystem or environment access; suitable for
//     hermetic builds and tests.
//   - Robust: never panics on malformed input; produces error tokens with
//     stable positions and messages to support editor tooling and E2E tests.
package scanner
