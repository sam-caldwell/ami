# Workspace Package/Manifest Audit — Implementation Summary

This document describes package‑level tooling added to support dependency management workflows without modifying the
`ami build` command (owned by Agent D).

## Scope
- Package/workspace features only (no changes to parser/lexer/compiler/build).
- Deterministic helpers with one‑concept‑per‑file organization and paired tests.

## Added Capabilities
- Version constraints for imports
  - File: `src/ami/workspace/constraints.go`
  - Supports exact (`v1.2.3`), `^`, `~`, `>`, `>=`, and `==latest`.
  - Functions: `ParseConstraint`, `Satisfies`.

- Import normalization and parsing
  - File: `src/ami/workspace/imports.go`
  - `NormalizeImports(*Package)`: trims + de‑dups while preserving order.
  - `ParseImportEntry(string) (path, constraint)`: splits entry into path and raw constraint.

- Workspace schema helpers (library only)
  - File: `src/ami/workspace/validate.go`
  - Validates SemVer, `toolchain.compiler.target`, `os/arch` format, and concurrency semantics. Not wired to CLI.

- Env dedup on load
  - File: `src/ami/workspace/load.go` (enhanced)
  - De‑duplicates `toolchain.compiler.env` preserving first occurrence order.

- Manifest helpers for `ami.sum`
  - File: `src/ami/workspace/manifest.go` (enhanced)
  - `Has`, `Set`, `Versions` for deterministic updates and queries.

- Requirements collection and cross‑checks
  - File: `src/ami/workspace/requirements.go`
  - `CollectRemoteRequirements(*Workspace)`: extracts remote (non‑`./`) import requirements; defaults to `==latest`
    when
    unspecified; collects parse errors separately.
  - `CrossCheckRequirements(*Manifest, []Requirement)`: reports names missing in `ami.sum` and
    present‑but‑unsatisfied.

- Cache/sum integrity helpers
  - File: `src/ami/workspace/sum_utils.go`
  - `UpdateManifestEntryFromCache`: computes sha256 via `HashDir` and records deterministically.
  - `CrossCheckRequirementsIntegrity`: intersects manifest integrity issues with versions satisfying constraints.
  - `DefaultCacheRoot`: resolves package cache path according to env/defaults.

- Orchestrator function (library only)
  - File: `src/ami/workspace/audit.go`
  - `AuditDependencies(dir string) (AuditReport, error)`: loads workspace and `ami.sum`, collects requirements,
    computes:
    - `MissingInSum`, `Unsatisfied` (content checks),
    - `MissingInCache`, `Mismatched` (integrity checks filtered to satisfying versions),
    - `ParseErrors`, `Requirements`, `SumFound` flags.

## Tests
All new functionality includes happy and sad path tests under `src/ami/workspace/*_test.go`:
- Constraints: accepted forms, rejections, and satisfaction rules.
- Imports: normalization and entry parsing.
- Validation: target path, env patterns, concurrency, defaults.
- Env dedup: preserves order while removing duplicates.
- Manifest helpers: set/has/versions behavior.
- Requirements cross‑checks: missing and unsatisfied scenarios.
- Sum utils: cache update and integrity filtering.
- Orchestrator: end‑to‑end summary with and without `ami.sum` present.

## Notes
- No edits to `ami build`, parser, lexer, or compiler code to avoid scope overlap with Agent D.
- Deterministic output patterns preserved (sorted keys, stable JSON in existing code paths).
- These helpers are intended for future CLI and build integrations while remaining library‑only for now.

## See Also
- `docs/cmd/mod.md` — user-facing module commands (`mod audit`, `mod update`, `mod sum`, etc.) that consume these
helpers.
