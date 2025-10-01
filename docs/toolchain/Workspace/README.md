# `ami.workspace` Schema (Workspace File)

This document describes the `ami.workspace` YAML file used by the `ami` toolchain to configure a workspace.

## Overview
- Location: workspace root as `ami.workspace`.
- Format: YAML
- Purpose: declare project metadata, toolchain settings, and packages.

## Top-level Keys
- `version` (string): schema version in SemVer format (e.g., `1.0.0`).
- `toolchain` (object): toolchain configuration groups.
- `packages` (list): list of workspace packages.

## Project Metadata
Although not currently enforced by the CLI, a conventional top-level `project` object may be provided:
- `project.name` (string)
- `project.version` (SemVer)

## Toolchain
- `toolchain.compiler` (object)
  - `concurrency` (string): either `NUM_CPU` or a positive integer as a string (e.g., `"4"`).
  - `target` (string): workspace-relative output directory (default `./build`). Must not be absolute or traverse outside the workspace.
  - `env` (list[string]): cross-compile targets in `os/arch` form (e.g., `linux/amd64`, `darwin/arm64`).
    - Accepted pattern: `^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$`
    - Known examples (extensible): `windows/amd64`, `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`.
    - Duplicates are eliminated when loading, preserving first occurrence order.
- `toolchain.linker` (object): reserved for future keys.
- `toolchain.linter` (object): reserved for future keys.

Notes:
- `ami init` seeds `toolchain.compiler.env` with the current host `GOOS/GOARCH` pair.

## Packages
The `packages` key uses a sequence of single-entry maps to preserve a logical key (e.g., `main`, `util`):

```yaml
packages:
  - main:
      name: app
      version: 0.1.0
      root: ./src
      import: ["./lib", "modA ^1.2.3", "modB >= 1.0.0"]
  - util:
      name: util
      version: 1.2.3
      root: ./util
      import: []
```

- `name` (string): package name.
- `version` (SemVer): `MAJOR.MINOR.PATCH` with optional prerelease (e.g., `1.2.3-rc.1`).
- `root` (string): workspace-relative path to the package root.
- `import` (list[string]): imports of other packages.
  - Local imports: paths beginning with `./`.
  - Remote imports: `module [constraint]` where constraint is optional.

## Version Constraints (imports)
Accepted forms:
- Exact: `X.Y.Z` (with optional leading `v`)
- Caret: `^X.Y.Z`
- Tilde: `~X.Y.Z`
- Greater-than: `>X.Y.Z`
- Greater-than-or-equal: `>=X.Y.Z`
- Macro latest: `==latest`

Rules:
- SemVer must be `MAJOR.MINOR.PATCH`; prereleases allowed when specified (e.g., `^1.0.0-rc.1`).
- Whitespace inside constraints is ignored (e.g., `>= 1.2.3`).
- Unsupported operators (e.g., `<=`) are rejected.

## Validation (library)
- SemVer checks for `version` and package versions.
- `toolchain.compiler.target` must be workspace-relative and must not escape the workspace.
- `toolchain.compiler.env` entries must match `os/arch` pattern. Duplicates are removed.

## Defaults
`ami init` creates a minimal workspace with:
- `version: 1.0.0`
- `toolchain.compiler.concurrency: NUM_CPU`
- `toolchain.compiler.target: ./build`
- `toolchain.compiler.env: [<host_os>/<host_arch>]`
- `packages` containing `main` with a starter package.

## Notes
- This document reflects the current implementation state in this repository. Future phases may evolve the shape (e.g., richer `toolchain.compiler.env`).
