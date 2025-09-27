# Complex Example

This example demonstrates a multi-package workspace:

- Cross-platform `toolchain.compiler.env` matrix
- Main package importing a local vendor package `alpha`
- Minimal pipeline to exercise parser/IR

Build it with the repo’s CLI:

1) Build the CLI: `go build -o ../../build/ami ../../src/cmd/ami`
2) From this directory: `../../build/ami build --verbose`

Outputs are written under `./build` and staged by `make examples` under `../../build/examples/complex/`.

