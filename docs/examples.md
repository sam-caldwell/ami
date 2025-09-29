# Examples

This repository includes example AMI workspaces under `examples/`.

- `examples/simple`: minimal single-package workspace with a trivial function and pipeline.
- `examples/complex`: multi-package workspace with a local vendor package.

## Building

1) Build the CLI: `go build -o build/ami ./src/cmd/ami` 2) Build all examples: `make examples`

The `examples` Makefile target iterates through `examples/*` and runs `ami build --verbose` for each directory
containing an `ami.workspace`. Resulting outputs are staged under `build/examples/<name>/`.

You can also build an individual example by changing into the example directory and running `../../build/ami build
--verbose`.

See also: `docs/make.md` for Make targets used by examples.
