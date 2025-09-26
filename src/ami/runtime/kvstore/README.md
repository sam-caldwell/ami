# AMI Runtime Key/Value Store (Go)

This package provides the key/value store used by the Go-based AMI runtime tester
("Phase 2" harness). The tester executes AMI Intermediate Representation (IR)
programs during pipeline simulation, so the kvstore is implemented in Go to
integrate directly with the rest of the harness, which is also written in Go.

## Why a Go Implementation Exists

* The AMI toolchain (compiler, runtime tester, and associated tooling) is a Go
  module. Implementing the kvstore in Go keeps the harness simple by reusing the
  Go ecosystem, build tooling, and standard library.
* Running the Go harness necessarily brings along Go's garbage-collected
  runtime. This is acceptable for the tooling layer because it is outside the
  target execution environment of real AMI programs.
* The current AMI backend produces textual AMI-IR assembly listings instead of
  native executables. Consequently, compiled AMI artifacts do not link against
  this Go code.

## Relationship to the Planned AMI Runtime

The AMI specification targets a runtime without automatic garbage collection and
with RAII-style resource management. When a production-quality AMI backend and
runtime are implemented, the kvstore (and other runtime services) will need to
be rewritten in a systems language that matches those constraints. The Go-based
kvstore serves as scaffolding for testing and validation during development, not
as the final runtime implementation.

## Implications for Contributors

* Changes here affect only the Go runtime tester harness.
* Do not rely on this package when designing the long-term AMI runtime API;
  treat it as tooling infrastructure.
* When work begins on the production runtime, plan to replace this package with
  an implementation that satisfies the no-GC, RAII expectations of AMI programs.
