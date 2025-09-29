# IR Indices (Debug)

The compiler emits per‑package IR indices in debug builds under `build/debug/ir/<pkg>/` to aid tooling and inspection.

Artifacts:
- `ir.index.json` — per‑package index of functions by unit
- `ir.types.index.json` — per‑unit set of type names used (params, results, and instruction operands/results)
- `ir.symbols.index.json` — per‑unit exported function names and referenced runtime externs

Schemas:

- `ir.index.json`
  - `schema`: `ir.index.v1`
  - `package`: package name
  - `units`: `[{"unit":"<name>", "functions":["F","G",...]}]`

- `ir.types.index.json`
  - `schema`: `ir.types.index.v1`
  - `package`: package name
  - `units`: `[{"unit":"<name>", "types":["int","bool",...]}]`

- `ir.symbols.index.json`
  - `schema`: `ir.symbols.index.v1`
  - `package`: package name
  - `units`: `[{"unit":"<name>", "exports":["F"], "externs":["ami_rt_alloc","ami_rt_panic"]}]`

Determinism:
- Function, type, and symbol lists are sorted; file paths are workspace‑relative where surfaced (e.g., in build plan).

References:
- Debug build manifest `build/debug/manifest.json` contains `irIndex`, `irTypesIndex`, and `irSymbolsIndex` paths for each package.
- The verbose build plan (`build/debug/build.plan.json`) includes `irIndex`, `irTypesIndex`, and `irSymbolsIndex` arrays when present.

## See Also
- `docs/backend/build_plan.md` — build plan schema that references these indices.
- `docs/edges.md` — edges index (`edges.v1`) written alongside other debug artifacts.
