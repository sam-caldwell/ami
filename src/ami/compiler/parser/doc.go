// Package parser builds ASTs from tokens with error recovery.
//
// Responsibilities
//
//   - Parse declarations, statements, and expressions from the scanner's token
//     stream into well-formed AST nodes.
//   - Implement precedence/associativity for operators and disambiguate
//     constructs according to the language grammar.
//   - Attach precise source positions to nodes to enable high-quality
//     diagnostics and editor tooling.
//
// Error Handling
//
// The parser reports errors with stable diagnostics and attempts local recovery
// to continue producing a useful AST for subsequent semantic analysis. Recovery
// should avoid cascading failures by synchronizing at statement and declaration
// boundaries.
//
// Invariants
//
//   - Deterministic: identical token input yields identical ASTs and
//     diagnostics ordering.
//   - No side effects: no filesystem or environment access; suitable for
//     hermetic builds and tests.
package parser
