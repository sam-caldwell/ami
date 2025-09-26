# CONFLICTS / ACTION ITEMS (docx vs. code)

The authoritative source is `docs/Asynchronous Machine Interface.docx`. This checklist captures divergences between 
the current implementation in `src/` and the grammar/semantics used across `docs/` and `examples/**`. Each item is an 
action to bring code and docs into alignment with the .docx and the updated examples.

## Grammar & Parser

- [x] Package version syntax: accept `package <name>:<version>` and record version in AST/IR. Ensure documentation matches implementation and that both are consistent with docs/Asynchronous Machine Interface.docx
- [x] Import lines with version constraints: accept `import <module> >= vX.Y.Z` (and block form), represent in AST, and surface in sources/debug artifacts.
- [x] Node attributes as structured key/value pairs: parse `in=...`, `worker=...`, `minWorkers`, `maxWorkers`, `onError`, `capabilities`, `type` as named attributes (not opaque strings).
- [x] Inline worker function literals: parse `worker=func(...) { ... }` into AST (function literal) and surface in AST/debug; IR debug propagation complete; codegen pending.
- [x] Expression parsing: allow reserved `state` identifier (from KW_STATE) as an expression identifier so `state.get(...)`/`state.set(...)` parse correctly.
- [x] Generic‑like calls: tolerate `Event<uint64>(...)` and similar type‑parameterized constructs in expressions sufficiently for AST/semantics scaffolding.
- [x] Function type parameters: parsed with tolerant grammar; added `FuncDecl.TypeParams` and `TypeParam{Name, Constraint}`; minimal semantics (reject duplicate names); IR surfaces `{name,constraint}`.
- [x] SPEC §6.2 pipeline examples vs. examples/docx: examples updated to canonical `worker=`/`in=` attributes and `Fanout` casing; parser supports attribute lists; SPEC/docs aligned with docx for these items.

## Semantics & IR

- [ ] Worker signature rules: standardize all node worker function signatures to `func(ev Event<T>) (Event<U>,error)`.
- [x] Standardize on attribute‑driven worker resolution: prefer `worker=` (ref or inline) and prefer `in=` for edges over positional heuristics.
- [ ] `edge.MultiPath` for Collect:
  - [ ] Parser support for `edge.MultiPath(inputs=[...], merge=Sort(...))`.
  - [ ] Semantic validation: only valid on `Collect`; inputs[0] must be the default upstream edge; each entry must be a valid `EdgeSpecifier`; type compatibility across inputs.
  - [ ] IR extensions: encode MultiPath configuration (inputs, merge attributes) in `pipelines.v1` and `edges.v1` summaries.
- [ ] `edge.Pipeline` type safety across pipelines: ensure type flow checks across pipeline boundaries use declared `type` on edges; add diagnostics on mismatch.
- [ ] Backpressure policy set: SPEC/semantics currently accept `block|drop`; examples/docx use `dropOldest` (and sometimes `dropNewest`).  standardize on dropOldest/dropNewest.
- [ ] Worker state parameter vs. ambient `state`: SPEC §6.3 shows `st *State` in signatures; docx and repo rule (Memory Safety 2.3.2) remove raw pointers and use ambient `state.get/set/update/list`. Remove pointer forms from examples/spec and update analyzer to not require a `State` parameter.
 - [x] Type parameters: emit `E_DUP_TYPE_PARAM` for duplicate type parameter names; enforced in semantics and exposed via lint.
 - [x] IR debug: surface function `typeParams` in `ir.v1` for tooling.

## Codegen & Tooling

- [x] IR lowering of inline workers: include input/output payload types and origin (literal vs. reference) for debug.
- [x] Build debug artifacts: ensure AST/IR JSON includes new fields (package version, import constraints, node attrs, inline workers). MultiPath pending.
- [x] Linter: prefer structured attrs for workers/edges; recognize `worker=` and `in=`.
 - [x] Linter JSON: `E_DUP_TYPE_PARAM` included when duplicate type parameters are present.

## Documentation Updates

- [x] `docs/compiler_grammar.md`: formalize EBNF for:
  - [x] Package with version (`package ident ':' version`)
  - [x] Import with version constraints
  - [x] Node attribute lists (`key '=' Expr { ',' key '=' Expr }`)
  - [x] Inline function literals and their allowed forms per node (scaffold)
  - [x] `edge.MultiPath` shape (constraints outlined; lowering pending)
- [x] `docs/edges.md`: document expanded backpressure policies; examples aligned; MultiPath references kept (lowering pending).
- [x] `docs/merge.md`: status updated to in‑progress; attribute names/examples aligned.
- [x] `SPECIFICATION.md`: updated for attribute grammar and ambient state; added remaining‑work items for MultiPath.
  - [x] SPEC §6.2: move toward attribute-form examples and `Fanout` casing (full sweep pending).
  - [x] SPEC §6.3: aligned with ambient `state`; removed raw pointer expectations.

## Tests

- [x] Parser tests for: package version, import constraints, attribute lists, inline workers, `state.*` usage, generic-like calls.
  - [ ] Parser/IR tests for `edge.MultiPath` pending (after lowering).
- [x] Semantics: duplicate function type parameter names.
  - [ ] Additional: worker signature variants, backpressure policy validation, MultiPath checks.
- [x] IR tests: inline worker metadata; function type params (with constraints) round-trip.
  - [ ] IR tests: MultiPath inputs/merge pending (after scaffold).
- [x] Linter: pipeline smells prefer structured attrs; existing tests cover no-workers and unbounded/block warnings.
 - [x] Linter JSON: duplicate type params test asserts `E_DUP_TYPE_PARAM` appears in `ami --json lint` output.
 - [x] Pipelines debug IR: extended test asserts step `Attrs` include `worker`, `minWorkers`, `maxWorkers`, `onError`, `capabilities`, and that `in=edge.*` (FIFO/LIFO/Pipeline) appears both in raw attrs and structured `inEdge`.

## Known SPEC vs Code Conflicts (Quick Index)

- [x] SPEC §6.2 casing: updated to `Fanout` to match examples and docx.
- [x] SPEC §6.2 pipeline examples: updated to named attribute lists (`worker=`, `in=`) instead of positional forms.
- [ ] SPEC §6.3 shows `st *State` parameter; docx uses ambient `state` and prohibits raw pointers.
- [ ] Backpressure tokens: SPEC/semantics `block|drop` vs examples/docx `dropOldest`/`dropNewest`.
- [ ] `edge.MultiPath`: SPEC marked planned, examples/docx demonstrate usage; parser/semantics not fully implemented.
- [x] Function type parameter lists (`func f<T,...>(...)`) are parsed; minimal semantics implemented (duplicate-name rejection). Broader generic constraint/unification remains pending.

## Notes

- The examples now serve as living grammar fixtures. Parser/semantics work should be validated directly against `examples/correct` and `examples/complex`.
- Treat `.docx` as source of truth for any ambiguity; update `.md` docs accordingly as part of the above items.
