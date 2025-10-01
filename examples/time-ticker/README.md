# Example: time-ticker

A minimal AMI workspace that uses the stdlib `time` module and defines a trivial pipeline so the compiler emits IR/ASM and other debug artifacts.

- Workspace: `examples/time-ticker/ami.workspace`
- Source: `examples/time-ticker/src/main.ami`

Build
- Build all examples: `make examples`
- Build only this example: `make example-time-ticker`

Outputs
- Debug artifacts (IR/ASM, etc.) are written under `examples/time-ticker/build/` and staged to `build/examples/time-ticker/`.

Notes
- The AMI `time` module provides `time.now`, `time.sleep`, `time.add`, and `time.delta` along with a `Ticker` facility.
- This example focuses on deterministic, reproducible outputs with `--verbose` enabled.

