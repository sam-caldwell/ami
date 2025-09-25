# Examples

This repo includes two example workspaces that you can build with the current `ami` CLI. These examples demonstrate the build pipeline and debug artifacts; runtime execution is not implemented yet (ingress/egress workers are stubs for now).

## Prerequisites

- Go 1.25+

## Build the CLI

- From repo root:
  - `go build -o build/ami ./src/cmd/ami`

Or use the Makefile target:

- `make build` (builds the CLI into `build/ami`)

## Simple Example

- Workspace: `examples/simple`
- Build (human):
  - `cd examples/simple`
  - `../../build/ami build`
- Build with debug artifacts:
  - `../../build/ami build --verbose`
- Outputs:
  - Non‑debug: `build/obj/<package>/<unit>.s` and per‑package index `build/obj/<package>/index.json`
  - Debug (`--verbose`):
    - `build/debug/source/resolved.json`
    - `build/debug/ast/<pkg>/<unit>.ast.json`
    - `build/debug/ir/<pkg>/<unit>.ir.json`
    - `build/debug/ir/<pkg>/<unit>.pipelines.json`
    - `build/debug/asm/<pkg>/<unit>.s` and `build/debug/asm/<pkg>/index.json`
  - `ami.manifest` is written at workspace root with artifact metadata.

## Complex Example

- Workspace: `examples/complex`
- Description:
  - Two timer‑driven ingress pipelines (1s current time, 2s incrementing counter) each transform to `uint(1)`.
  - A third pipeline collects/merges both streams, counts events, and egresses the running count to stdout (at runtime; not executed by build).
- Build:
  - `cd examples/complex`
  - `../../build/ami build --verbose`
- Inspect debug outputs (examples):
  - `build/debug/ir/main/main.ami.pipelines.json`
  - `build/debug/asm/main/edges.json`

## Notes

- The build validates `ami.workspace`; if `ami.sum` is present, integrity is checked against the local cache.
- Debug artifacts are only emitted with `--verbose`.
- The examples do not require any external modules; no `ami mod` steps are needed.

## Makefile Convenience

- From repo root, run `make examples` to build all example workspaces and collect their outputs under `build/examples/<name>/`.
  - Requires `build/ami` to exist (run `make build` first if needed).
