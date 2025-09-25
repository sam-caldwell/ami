# AMI Stdlib Determinism Policy (Phase 1)

Authority and Scope

- The authoritative source for POP/AMI language and runtime semantics is `docs/Asynchronous Machine Interface.docx`.
- This document summarizes determinism requirements for the AMI standard library (stdlib) and is normative only insofar as it does not conflict with the .docx. If any conflict arises, the .docx prevails and this document must be updated.

Core Principles

- Explicit Inputs → Stable Outputs: For identical inputs and configuration, functions produce identical outputs across OS/arch and runs.
- No Ambient State: No implicit use of process environment, global RNGs, system time, locale, or filesystem state.
- Explicit Effects: All I/O and side effects are explicit through function parameters and return values; no background work or hidden retries.
- Canonical Forms: Where ordering matters (e.g., JSON object keys), outputs must be canonical and reproducible.
- Isolation: Stdlib functions do not reach outside their scope (e.g., no network) unless explicitly designed and documented for I/O.

Time and Randomness

- Time: "Now" can only be obtained via an injected `Clock` interface; formatting/parsing uses ISO‑8601 UTC with milliseconds. No direct system clock reads.
- Randomness: All PRNGs require explicit seeds or injected RNG interfaces. No package‑level global RNGs, and no OS entropy sources.

Concurrency and Scheduling

- No Hidden Concurrency: Stdlib must not create implicit goroutines. Any concurrency must be explicit and expose deterministic ordering/limits.
- Deterministic Buffers: Buffered I/O sizes and flush semantics must be specified and reproducible.

Filesystem and Paths

- Pure String Transforms: `path/filepath` utilities operate as pure transforms (no symlink evaluation, no cwd dependence).
- Explicit File I/O: `os` functions for Read/Write/Mkdir/Stat are explicit; permissions and modes are parameters, not global defaults.

Encoding and Data Formats

- JSON: Map/object key ordering must be stable and canonical. Provide strict decode options; document tag handling.
- Bytes/Strings: Unicode and byte operations must be consistent across platforms; document any edge cases explicitly.

Cryptography (Phase 1 Scope: SHA‑256 only)

- Deterministic Hashing: One‑shot and streaming SHA‑256 must produce stable results using published test vectors. No ambient randomness.
- Constant‑Time Considerations: For future crypto packages, operations must be constant‑time where applicable and avoid side‑channel leaks.

Errors and Diagnostics

- Stable Errors: Error kinds and messages must be documented and consistent; avoid embedding non‑deterministic data (timestamps, pointers) in error strings.
- No Panics for Expected Conditions: Return errors rather than panicking for user‑reachable conditions.

Memory Safety (AMI 2.3.2 Alignment)

- No Raw Pointers: AMI does not expose raw pointers or process addresses. Stdlib APIs must not require or expose pointer semantics to AMI programs.
- Assignment Semantics: The `*` operator in AMI marks mutating assignment and is not a dereference; no `&` operator. Stdlib design and documentation must not suggest pointer‑based mutation.

Testing and Coverage

- Golden Tests: Provide golden tests for canonicalization (JSON ordering, ISO‑8601 UTC formatting, filepath normalization).
- Sad Paths: Every package includes negative tests for invalid inputs, boundary conditions, and policy denials (e.g., catastrophic regex patterns).
- Coverage: Target ≥80% coverage per package; minimum 75%.

Platform Stability

- Cross‑Platform Behavior: Document and test Windows/Unix differences (path separators, newlines) and normalize where required to maintain determinism.

Documentation and Versioning

- Per‑Package Docs: Each stdlib package has a `docs/stdlib/<pkg>.md` describing API, determinism guarantees, and examples.
- Change Management: Any behavioral change affecting determinism must be documented and accompanied by updated tests.

Compliance Note

- Where this policy is silent, implementers must adhere to the spirit of determinism, explicit effects, and the constraints defined in `docs/Asynchronous Machine Interface.docx` (authoritative).

