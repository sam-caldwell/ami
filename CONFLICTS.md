# CONFLICTS / ACTION ITEMS (docx vs. code)

The authoritative source is `docs/Asynchronous Machine Interface.docx`. This checklist captures divergences between 
the current implementation in `src/` and the grammar/semantics used across `docs/` and `examples/**`. Each item is an 
action to bring code and docs into alignment with the .docx and the updated examples.

## Grammar & Parser

- [x] Package version syntax: accept `package <name>:<version>` and record version in AST/IR. Ensure documentation matches implementation and that both are consistent with docs/Asynchronous Machine Interface.docx
- [ ] Import lines with version constraints: accept `import <module> >= vX.Y.Z` (and block form), represent in AST, and surface in sources/debug artifacts.
- [ ] Node attributes as structured key/value pairs: parse `in=...`, `worker=...`, `minWorkers`, `maxWorkers`, `onError`, `capabilities`, `type` as named attributes (not opaque strings).
- [ ] Inline worker function literals: parse `worker=func(...) { ... }` into AST (function literal) and propagate into IR/codegen.
- [ ] Expression parsing: allow reserved `state` identifier (from KW_STATE) as an expression identifier so `state.get(...)`/`state.set(...)` parse correctly.
- [ ] Generic‑like calls: tolerate `Event<uint64>(...)` and similar type‑parameterized constructs in expressions sufficiently for AST/semantics scaffolding.
- [ ] Function type parameters: examples demonstrate function type parameter lists (e.g., `func ingressWorker<T []byte, N any>(...) ...)`); the parser does not accept this form. Define grammar + AST nodes for function type parameters and constraints (`any`, concrete), or align examples/spec if out of scope.
- [ ] SPEC §6.2 pipeline examples vs. examples/docx: SPEC still shows legacy positional forms `Ingress(cfg).Transform(f).FanOut(a,b).Collect().Egress(cfg)` and `FanOut` casing. Canonicalize on attribute lists (e.g., `worker=`, `in=`) and `Fanout` casing, and reflect this in SPEC examples.

## Semantics & IR

- [ ] Worker signature rules: extend to accept node‑specific simplified forms used in examples (e.g., Transform worker `func(ev Event<T>) Event<U>`). Keep support for context/state forms where needed.
- [ ] Attribute‑driven worker resolution: read `worker=` from structured attrs (ref or inline func) instead of positional arg heuristics; stop treating `in=` as a “worker name”.
- [ ] `edge.MultiPath` for Collect:
  - [ ] Parser support for `edge.MultiPath(inputs=[...], merge=Sort(...))`.
  - [ ] Semantic validation: only valid on `Collect`; inputs[0] must be the default upstream edge; each entry must be a valid `EdgeSpecifier`; type compatibility across inputs.
  - [ ] IR extensions: encode MultiPath configuration (inputs, merge attributes) in `pipelines.v1` and `edges.v1` summaries.
- [ ] `edge.Pipeline` type safety across pipelines: ensure type flow checks across pipeline boundaries use declared `type` on edges; add diagnostics on mismatch.
- [ ] Backpressure policy set: SPEC/semantics currently accept `block|drop`; examples/docx use `dropOldest` (and sometimes `dropNewest`). Decide canonical tokens and extend parser/semantics; continue mapping to `atLeastOnce|bestEffort` deterministically.
- [ ] Worker state parameter vs. ambient `state`: SPEC §6.3 shows `st *State` in signatures; docx and repo rule (Memory Safety 2.3.2) remove raw pointers and use ambient `state.get/set/update/list`. Remove pointer forms from examples/spec and update analyzer to not require a `State` parameter.

## Codegen & Tooling

- [ ] IR lowering of inline workers: include input/output payload types and origin (literal vs. reference) for codegen/debug.
- [ ] Build debug artifacts: ensure AST/IR JSON includes new fields (package version, import constraints, node attrs, inline workers, MultiPath).
- [ ] Linter: update worker/edge detection so `analyzeWorkers` does not misinterpret `in=` attributes as worker calls; recognize `worker=` specifically.

## Documentation Updates

- [ ] `docs/compiler_grammar.md`: formalize EBNF for:
  - [ ] Package with version (`package ident ':' version`)
  - [ ] Import with version constraints
  - [ ] Node attribute lists (`key '=' Expr { ',' key '=' Expr }`)
  - [ ] Inline function literals and their allowed forms per node
  - [ ] `edge.MultiPath` shape and constraints
- [ ] `docs/edges.md`: document expanded backpressure policies and complete `edge.MultiPath` examples to match the .docx and examples.
- [ ] `docs/merge.md`: reconcile status vs. examples; mark planned → in‑progress; align attribute names and examples with `examples/**`.
- [ ] `SPECIFICATION.md`: update 6.6/6.7 checklists to reflect parser/IR/semantics tasks for MultiPath and attribute grammar; add remaining‑work checkboxes.
  - [ ] SPEC §6.2: replace positional `cfg`/`f` forms with attribute lists and correct `FanOut`→`Fanout` casing; include `edge.MultiPath` usage in Collect examples.
  - [ ] SPEC §6.3: remove `*State` from worker signatures; document ambient `state` per docx §2.2.14/2.3.5 and repository Memory Safety note.

## Tests

- [ ] Parser golden tests for: package version, import constraints, attribute lists, inline workers, `state.*` usage, `edge.MultiPath`.
- [ ] Semantics tests for worker signature variants (per node), backpressure policy validation, MultiPath context/arity/type checks.
- [ ] IR/codegen tests: capture inline worker metadata, MultiPath inputs/merge in `pipelines.v1` and `edges.v1`.
- [ ] Linter tests: ensure no false positives on `in=`; detect missing/invalid `worker=`.

## Known SPEC vs Code Conflicts (Quick Index)

- [ ] SPEC §6.2 uses `FanOut` casing; examples/docs use `Fanout`.
- [ ] SPEC §6.2 pipeline example uses positional `cfg`/`f`; examples/docs/docx use named attribute lists.
- [ ] SPEC §6.3 shows `st *State` parameter; docx uses ambient `state` and prohibits raw pointers.
- [ ] Backpressure tokens: SPEC/semantics `block|drop` vs examples/docx `dropOldest`/`dropNewest`.
- [ ] `edge.MultiPath`: SPEC marked planned, examples/docx demonstrate usage; parser/semantics not fully implemented.
- [ ] Function type parameter lists (`func f<T,...>(...)`) appear in examples; parser/type system do not accept them today.

## Notes

- The examples now serve as living grammar fixtures. Parser/semantics work should be validated directly against `examples/correct` and `examples/complex`.
- Treat `.docx` as source of truth for any ambiguity; update `.md` docs accordingly as part of the above items.
