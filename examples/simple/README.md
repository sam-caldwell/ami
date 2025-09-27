# Simple Example

This example demonstrates a minimal AMI workspace:

- Cross-platform `toolchain.compiler.env` matrix
- A trivial function and a minimal pipeline

Build it with the repo’s CLI:

1) Build the CLI: `go build -o ../../build/ami ../../src/cmd/ami`
2) From this directory: `../../build/ami build --verbose`

Outputs are written under `./build` and staged by `make examples` under `../../build/examples/simple/`.

