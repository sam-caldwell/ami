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
  - [x] Repo formatting baseline: applied one‑declaration‑per‑file to struct AST types (`StructDecl`, `Field`), aligning with token package pattern.
 - [x] Function‑body comments: attach leading comments to body statements (`VarDeclStmt`, `ExprStmt`, `AssignStmt`, `DeferStmt`, `ReturnStmt`) via token‑offset map; covered by `parser/body_comments_test.go`.

## Semantics & IR

- [ ] Worker signature rules: standardize all node worker function signatures to `func(ev Event<T>) (Event<U>, error)`.
  - Status: analyzer now accepts the canonical ambient form and continues to tolerate legacy `(Context, Event<T>, State)` during migration. Result kinds for legacy remain `Event<U> | []Event<U> | Error<E>`; canonical uses `(Event<U>, error)`.
  - Impact: pipelines/tests previously assumed broader forms; IR still captures `HasContext/HasState` and `OutputKind` for compatibility during transition.
  - Plan:
    - Enforce canonical ambient signature by default; keep a deprecation window for explicit `State` parameters.
    - Maintain clear diagnostics with positions (`E_WORKER_SIGNATURE`).
    - Deprecate IR `HasState/HasContext` after downstream tooling is updated.
    - Migrate examples/tests to ambient signature; retain explicit tests for deprecation path.
  - Progress:
    - [x] `E_WORKER_SIGNATURE` includes positions (lint/build JSON tests updated).
    - [x] `E_WORKER_UNDEFINED` includes positions.
    - [x] Canonical acceptance: `func(ev Event<T>) (Event<U>, error)` recognized by semantics; legacy remains accepted.
    - [x] Deprecations:
      - Lint: `W_STATE_PARAM_AMBIENT_SUGGEST` for explicit `State` parameters (info).
      - Semantics: `W_WORKER_STATE_PARAM_DEPRECATED` when a legacy worker is referenced (info).
    - [x] IR: derives Input/Output from canonical signature (ignores trailing `error`).
    - [ ] Future: escalate legacy to warn/error and remove `HasContext/HasState` from IR; migrate all docs/examples.
- [x] Standardize on attribute‑driven worker resolution: prefer `worker=` (ref or inline) and prefer `in=` for edges over positional heuristics.
- [x] `edge.MultiPath` for Collect:
  - [x] Parser tolerant consumption via attribute string; IR scaffold parses `edge.MultiPath(inputs=[...], merge=Sort(...))`.
  - [x] Semantic validation (minimal): only valid on `Collect`; inputs required; first input must be a default upstream edge (FIFO); enforce input type compatibility; basic merge op shape (allowed names, parentheses).
  - [x] IR extensions: encode MultiPath configuration (inputs, merge attributes) in `pipelines.v1`. Edges summary now also carries a MultiPath scaffold for debug parity.
- [x] `edge.Pipeline` type safety across pipelines: ensure type flow checks across pipeline boundaries use declared `type` on edges; add diagnostics on mismatch.
  - Status: Implemented via `analyzeEdgePipelineTypeSafety` and wired into `AnalyzeFile`.
  - Behavior: Builds a pipeline→output payload type map by inspecting the penultimate step (before `Egress`) worker outputs; validates `edge.Pipeline(name=...,type=...)` against inferred type.
  - Diagnostics: `E_EDGE_PIPE_NOT_FOUND`, `E_EDGE_PIPE_TYPE_MISMATCH`.
  - Tests: Added happy/sad/unknown cases in `src/ami/compiler/sem/edge_pipeline_typesafety_test.go`.
  - Notes: If a pipeline’s output type cannot be inferred or is inconsistent across workers, cross‑pipeline mismatch is not emitted (conservative).
- [ ] Backpressure policy set: SPEC/semantics currently accept `block|drop`; examples/docx use `dropOldest` (and sometimes `dropNewest`).  standardize on dropOldest/dropNewest.
  - Status: compiler/semantics currently accept `block`, `dropOldest`, and `dropNewest` for FIFO/LIFO/Pipeline edges; tests and examples already use `block` and `dropOldest`/`dropNewest`. No bare `drop` token is accepted in code today.
  - Plan:
    - Documentation: align SPEC/examples to the canonical set `{ block, dropOldest, dropNewest }`; remove any lingering `drop` references.
    - Compatibility: if `drop` appears in external content, consider a transitional warning `W_EDGE_BP_ALIAS` mapping `drop -> dropOldest` (or reject outright per docx).
    - Tests: sweep for `drop` and update; add a sad‑path diagnostic for unknown backpressure tokens.
  - Progress:
    - [x] Semantics/code/tests use canonical `dropOldest`/`dropNewest`; bare `drop` is not accepted.
    - [ ] Docs/SPEC sweep to remove `drop` and adopt canonical terms.
    - [x] Linter: added `W_EDGE_BP_ALIAS` when encountering `drop` with a migration hint; test added.

## Linter

- [x] MultiPath shape and policy hints:
  - `W_MP_ONLY_COLLECT`, `W_MP_INVALID`, `W_MP_INPUTS_EMPTY`, `W_MP_INPUT0_KIND`.
  - Input smells: `W_MP_EDGE_SMELL_UNBOUNDED_BLOCK`, `W_MP_EDGE_SMELL_TINY_BOUNDED_DROP`.
  - Type mismatch hint: `W_MP_INPUT_TYPE_MISMATCH`.
  - Merge hints: `W_MP_MERGE_SUGGEST` (absent), `W_MP_MERGE_INVALID` (unknown names).

## Codegen & Tooling

- [x] MultiPath codegen scaffolding (no-op):
  - GenerateASM emits `mp_begin/mp_input/mp_merge/mp_end` pseudo-ops for steps using `edge.MultiPath(...)` to aid future integration and testing.
  - Existing single-edge `edge_init` pseudo-ops remain unchanged.
- [ ] MultiPath codegen mapping (future):
  - Lower `edge.MultiPath` to runtime merge orchestration with deterministic buffering and policy handling.
  - Map merge operators (`Sort`, `Stable`, `Key`, `Dedup`, `Window`, `Watermark`, `Timeout`, `Buffer`, `PartitionBy`) to concrete strategies.
- [x] Worker state parameter vs. ambient `state`: SPEC §6.3 shows `st *State` in signatures; docx and repo rule (Memory Safety 2.3.2) remove raw pointers and use ambient `state.get/set/update/list`. Remove pointer forms from examples/spec and update analyzer to not require a `State` parameter.
  - Status: parser/semantics reject pointer `*State` parameters (emits `E_STATE_PARAM_POINTER`). Non‑pointer `State` parameters are tolerated during migration; ambient `state.*` helpers are used in docs/tests. Linter warns on `*State` (W_STATE_PARAM_POINTER) and suggests ambient access for explicit `State` parameters (W_STATE_PARAM_AMBIENT_SUGGEST). Canonical signature migration is planned.
  - Plan:
    - Enforce canonical worker signature without explicit `State` param once finalized (escalate transitional notice to warn/error).
    - Linter: add guidance to remove explicit `State` parameters (e.g., `W_STATE_PARAM_AMBIENT_SUGGEST`) and flag pointer forms at parse or lint time.
    - IR: remove `HasState` after migration; adjust IR lowering/tests accordingly.
    - Docs/Examples: sweep to remove pointer/non‑ambient forms; prefer `state.get/set/update/list`.
  - Progress:
    - [x] Parser/semantics error `E_STATE_PARAM_POINTER` for pointer state params.
    - [x] Transitional notice added via `W_WORKER_SIGNATURE_DEPRECATED` (info) for legacy (Context, Event<T>, State) signatures.
    - [ ] Linter guidance for non‑pointer `State` parameters.
    - [ ] IR cleanup and docs/examples sweep.
  - Status: pointers are prohibited (no `&`, unary `*` non‑dereference). Many tests still demonstrate worker signatures with an explicit `State` parameter (non‑pointer); IR records `HasState` accordingly. Ambient `state.*` helpers are supported in docs and tests.
  - Plan:
    - Documentation/examples: remove `*State` forms; prefer ambient `state.*` access patterns.
    - Semantics: when worker signature enforcement lands, do not require an explicit `State` parameter; treat state as ambient by default. Provide a deprecation warning if an explicit `State` parameter remains (e.g., `W_WORKER_STATE_PARAM_DEPRECATED`).
    - IR: deprecate `HasState` once enforcement migrates to ambient‑only model; keep temporarily for debug compatibility.
    - Tests: migrate worker examples to ambient state; update analyzer tests accordingly.
 - [x] Type parameters: emit `E_DUP_TYPE_PARAM` for duplicate type parameter names; enforced in semantics and exposed via lint.
 - [x] IR debug: surface function `typeParams` in `ir.v1` for tooling.
 - [x] Diagnostics positions: `E_DUP_TYPE_PARAM` includes a source position (function start). Per‑parameter offsets are a future enhancement.

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
  - [x] Additional:
    - [x] Worker signature variants: invalid/canonical/legacy across attribute and positional refs (lint/build diagnostics).
    - [x] Backpressure policy validation: FIFO/LIFO/Pipeline (min/max order, policy set) with sad-path coverage.
    - [x] MultiPath checks: Collect-only, first input FIFO required, input type compatibility, merge operator validation.
- [x] IR tests: inline worker metadata; function type params (with constraints) round-trip.
  - [x] IR tests: MultiPath inputs/merge covered (see `src/ami/compiler/ir/multipath_schema_test.go` verifying inputs array and merge ops in `pipelines.v1`).
- [x] Linter: pipeline smells prefer structured attrs; existing tests cover no-workers and unbounded/block warnings.
- [x] Linter JSON: duplicate type params test asserts `E_DUP_TYPE_PARAM` appears in `ami --json lint` output.
- [x] Pipelines debug IR: extended test asserts step `Attrs` include `worker`, `minWorkers`, `maxWorkers`, `onError`, `capabilities`, and that `in=edge.*` (FIFO/LIFO/Pipeline) appears both in raw attrs and structured `inEdge`.
 - [x] Lint JSON: ambient state migration hint (`W_STATE_PARAM_AMBIENT_SUGGEST`) test added; semantics emit `W_WORKER_STATE_PARAM_DEPRECATED` on legacy worker use.
 - [x] Lint/Build JSON: `E_WORKER_SIGNATURE` includes `pos` (position) tests added.
 - [x] Lint JSON: state pointer (`W_STATE_PARAM_POINTER`) and ambient suggestion (`W_STATE_PARAM_AMBIENT_SUGGEST`) tests added.

## Known SPEC vs Code Conflicts (Quick Index)

- [x] SPEC §6.2 casing: updated to `Fanout` to match examples and docx.
- [x] SPEC §6.2 pipeline examples: updated to named attribute lists (`worker=`, `in=`) instead of positional forms.
- [x] SPEC §6.3 shows `st *State` parameter; docx uses ambient `state` and prohibits raw pointers.
  - Follow-up alignment completed:
    - [x] Parser/Semantics reject `*State` (`E_STATE_PARAM_POINTER`).
    - [x] Lint emits ambient migration hint for explicit `State` (`W_STATE_PARAM_AMBIENT_SUGGEST`).
    - [x] Semantics emit legacy worker deprecation when referenced (`W_WORKER_STATE_PARAM_DEPRECATED`).
    - [ ] Full docs/examples sweep to replace remaining `st State` examples with ambient `state.*` access.
- [ ] Backpressure tokens: SPEC/semantics `block|drop` vs examples/docx `dropOldest`/`dropNewest`.
  - Status: code/semantics accept `{ block, dropOldest, dropNewest }`; no bare `drop` token. Docs/SPEC still reference `drop`.
  - Action: update SPEC/docs to the canonical set; consider transitional alias warning if `drop` is encountered.
- [x] `edge.MultiPath`: Scaffold implemented (parser tolerance, minimal semantics, IR/schema + pseudo‑ops). Merge attribute normalization and runtime lowering pending.
- [x] Function type parameter lists (`func f<T,...>(...)`) are parsed; minimal semantics implemented (duplicate-name rejection). Broader generic constraint/unification remains pending.

## Next Steps for Remaining Conflicts

- Ambient state (SPEC §6.3)
  - [x] Parser/Semantics: reject pointer `*State` parameters (E_STATE_PARAM_POINTER).
  - [x] Linter: add rule to flag `*State` in function parameters (W_STATE_PARAM_POINTER).
  - [x] Linter: add suggestion to remove non‑pointer `State` params in favor of ambient access (W_STATE_PARAM_AMBIENT_SUGGEST).
  - [x] Docs/Examples: sweep to remove pointer forms; align with Memory Safety 2.3.2 guidance.

- Backpressure tokens alignment
  - [x] Parser/Semantics: accept `dropOldest`/`dropNewest` tokens and map deterministically to runtime policy.
  - [x] Linter: warn on legacy `drop` when ambiguous; provide migration hint. (W_EDGE_BP_AMBIGUOUS_DROP)
  - [x] Docs/Schemas/Examples: update to canonical terms and clarify delivery semantics (docs/edges.md, docs/merge.md).

- edge.MultiPath implementation
  - [x] Parser: finalize `in=edge.MultiPath(<attr list>)` on Collect; accept `merge.*` attributes (tolerant via raw attribute). Added parser round‑trip test.
  - [ ] Semantics: context checks (Collect‑only), attribute arity/type validation, conflicts, and required fields.
        Status: Minimal checks complete (Collect‑only, inputs required, FIFO first, type compatibility, allowed merge names). Deeper arity/type validation still pending.
  - [x] IR/Debug: emit MultiPath in `pipelines.v1` and `edges.v1`.
        Status: Scaffold emitted (inputs + merge ops {name, raw}). Normalized merge config (key/order/stable/window/etc.) pending.
  - [x] Lint/Tests: smells/hints for tiny buffers and missing fields; golden IR snapshots.
        Status: Lint hints for MultiPath present; added build test to assert edges.v1 MultiPath emission; existing IR test verifies pipelines.v1 scaffold.

## Notes

- The examples now serve as living grammar fixtures. Parser/semantics work should be validated directly against `examples/correct` and `examples/complex`.
- Treat `.docx` as source of truth for any ambiguity; update `.md` docs accordingly as part of the above items.
