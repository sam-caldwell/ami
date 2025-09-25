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
  - `build/debug/asm/<package>/<unit>.s`: assembly scaffold and per‑package index at `build/debug/asm/<package>/index.json`.
- Writes `ami.manifest` with artifact metadata and resolved packages (from `ami.sum`).

Notes:
- Debug artifacts are only produced with `--verbose`.
- Paths are stable and relative to the workspace.
