# ami build

Builds the workspace and (when `--verbose`) emits debug artifacts.

## Usage

- Human: `ami build`
- Verbose: `ami build --verbose`
- JSON: `ami --json build`

## Behavior

- Loads `ami.workspace` and prepares a build plan.
- With `--verbose`, writes debug artifacts under `build/debug/`:
  - `build/debug/source/resolved.json`: resolved sources list.
  - `build/debug/ast/<package>/<unit>.ast.json`: AST scaffold.
  - `build/debug/ir/<package>/<unit>.ir.json`: IR scaffold.
  - `build/debug/ir/<package>/<unit>.pipelines.json`: pipelines structure with worker references and generic payloads (debug IR for tooling).
  - `build/debug/ir/<package>/<unit>.eventmeta.json`: event lifecycle metadata contract (id, timestamp, attempt) and structured trace context (trace.traceparent/tracestate), with `immutablePayload=true`.
  - `build/debug/asm/<package>/<unit>.s`: assembly scaffold and per‑package index at `build/debug/asm/<package>/index.json`.
    - The index (asm.v1) includes an optional `edges` array mirroring `edges.json` for convenience.
  - Edge stubs: when pipelines declare `in=edge.*(...)`, the assembly includes `edge_init` pseudo‑ops (e.g., `edge_init label=P.step2.in kind=fifo min=10 max=20 bp=block type=[]byte`). In human verbose mode, these lines are also echoed to stdout to make high‑performance queue wiring visible during the build.
  - Edge summary: per‑package `build/debug/asm/<package>/edges.json` (`edges.v1`) lists all discovered input edges with pipeline, step, node, and parameters for quick inspection and tooling.
- Writes `ami.manifest` with artifact metadata and resolved packages (from `ami.sum`).

Notes:
- Debug artifacts are only produced with `--verbose`.
- Paths are stable and relative to the workspace.
