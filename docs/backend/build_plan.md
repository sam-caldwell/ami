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

- Paths under `objIndex` and `objects` are workspace‑relative.
- `hasObjects` is conservative and true when either real or stub `.o` files have
  been emitted. During early backend phases, stub `.o` files may exist even when
  the host toolchain is missing; when the toolchain is available, compiled
  objects will be preferred in indexes.

## Compatibility

- This schema is versioned via the `schema` field and may evolve. New optional
  fields will be added to preserve backward compatibility for existing
  consumers.

