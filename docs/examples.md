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

## POP Quickstart

The `examples/simple` workspace contains a minimal POP program that demonstrates:

- a trivial function and a minimal pipeline `P` with ingress/egress
- a workspace configured for multiple targets (darwin/linux/windows)

Run:

1) Build the CLI: `go build -o build/ami ./src/cmd/ami`
2) Change into the example directory: `cd examples/simple`
3) Build: `../../build/ami build --verbose`

This emits deterministic artifacts under `./build/` including per‑target LLVM/object files, a verbose debug tree (AST/IR/ASM), and `ami.manifest`.
