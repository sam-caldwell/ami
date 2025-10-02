// Package source provides file and position tracking for the compiler frontend.
//
// Responsibilities
//
//   - Manage file sets, line offsets, and byte ranges for mapping between
//     token positions and human-readable locations (file:line:column).
//   - Provide stable, comparable position values that can be stored on AST and
//     used by diagnostics throughout the compiler pipeline.
//   - Keep path and position handling portable and deterministic to support
//     reproducible builds and golden tests.
//
// Scope
//
// This package is intentionally focused on in-memory bookkeeping and does not
// perform filesystem I/O. It is consumed by the scanner, parser, and semantic
// analysis to attribute diagnostics and debug artifacts.
package source
