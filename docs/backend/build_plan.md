# Build Plan Schema (build.plan/v1)

This document describes the JSON schema and semantics of the debug build plan
written by `ami build --verbose` to `build/debug/build.plan.json`.

The plan is a human- and machine-consumable summary of what the builder
intends to do (targets, packages) and the artifacts that were discovered or
produced by the front-end and early backend stages.

## Top-Level Object

- `schema`: string. Always `"build.plan/v1"`.
- `targetDir`: string. Absolute path to the workspace target directory.
- `targets`: array of strings. OS/arch pairs (e.g., `"darwin/arm64"`).
- `packages`: array of package entries (see below).
- `objIndex` (optional): array of relative paths to per‑package object indexes
  (`build/obj/<pkg>/index.json`) when present.
- `objects` (optional): array of relative paths to discovered object files
  (`build/obj/<pkg>/*.o`) when present.
- `objectsByEnv` (optional): map of `env -> []string` listing per‑environment object files when cross‑env builds emit objects under `build/<env>/obj/...`.
- `objIndexByEnv` (optional): map of `env -> []string` listing per‑environment object index files under `build/<env>/obj/<pkg>/index.json`.
- `irIndex` (optional): array of relative paths to per‑package IR indices under `build/debug/ir/<pkg>/ir.index.json`.
- `irTypesIndex` (optional): array of relative paths to per‑package IR types indices under `build/debug/ir/<pkg>/ir.types.index.json`.
- `irSymbolsIndex` (optional): array of relative paths to per‑package IR symbols indices under `build/debug/ir/<pkg>/ir.symbols.index.json`.

Example:

```
{
  "schema": "build.plan/v1",
  "targetDir": "/abs/workspace/build",
  "targets": ["darwin/arm64"],
  "packages": [
    {
      "key": "main",
      "name": "app",
      "version": "0.0.1",
      "root": "./src",
      "hasObjects": true
    }
  ],
  "objIndex": [
    "build/obj/app/index.json"
  ],
  "objects": [
    "build/obj/app/u.o"
  ],
  "objectsByEnv": {
    "darwin/arm64": ["build/darwin/arm64/obj/app/u.o"],
    "linux/arm64": ["build/linux/arm64/obj/app/u.o"]
  },
  "objIndexByEnv": {
    "darwin/arm64": ["build/darwin/arm64/obj/app/index.json"]
  },
  "irIndex": [
    "build/debug/ir/app/ir.index.json"
  ],
  "irTypesIndex": [
    "build/debug/ir/app/ir.types.index.json"
  ],
  "irSymbolsIndex": [
    "build/debug/ir/app/ir.symbols.index.json"
  ]
}
```

## Package Entry

Each entry summarizes a logical package from `ami.workspace`.

- `key`: string. Workspace key (e.g., `"main"`).
- `name`: string. Package name (e.g., `"app"`).
- `version`: string. Package version.
- `root`: string. Workspace‑relative source root (e.g., `"./src"`).
- `hasObjects`: boolean. True if any `.o` objects exist under `build/obj/<pkg>/`.

## Semantics

- Paths under `objIndex`, `objects`, `objectsByEnv`, and IR index arrays are workspace‑relative.
- `hasObjects` is conservative and true when either real or stub `.o` files have been emitted. During early
  backend phases, stub `.o` files may exist even when the host toolchain is missing; when the toolchain is
  available, compiled objects will be preferred in indexes.

## Compatibility

- This schema is versioned via the `schema` field and may evolve. New optional fields may be added to preserve
  backward compatibility for existing consumers. Unknown optional fields should be ignored by readers.

## See Also
- `docs/ir-indices.md` — IR indices written under `build/debug/ir/<pkg>/...` referenced by `irIndex`, `irTypesIndex`, and `irSymbolsIndex`.
- `docs/edges.md` — edges index (`edges.v1`) emitted under `build/debug/asm/<pkg>/edges.json` during verbose builds.
- `docs/Workspace/README.md` — workspace file that configures targets and packages summarized in the plan.
