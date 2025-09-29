# Events, Errors, and Contracts (Beginner Guide)

This guide explains AMI’s event model, error typing, and contracts at a high level. It’s written for newcomers and links
back to the authoritative `.docx` spec.

Authoritative reference: docs/Asynchronous Machine Interface.docx (Ch. 1.7 and 2.2)

What is an “Event”?
- An event is a single unit of data flowing through a pipeline.
- Events are immutable: once created, the payload is not changed in place. New events are produced from existing events.
- Typical metadata: unique id, timestamp, attempt number, trace context (to follow flow across steps), and origin.

Event Typing
- Events have a payload type (what the data is) and metadata fields (who/when/where).
- Payload can be scalars (numbers, strings), containers (slice/set/map), or structured types defined by the program.
- Consistent typing enables static checks, better lints, and predictable behavior at runtime.
 - Advanced: Optional/Union types are supported in payloads where needed:
   - `Optional<T>` indicates a value may be absent; deep field resolution propagates optionality.
   - `Union<A,B,...>` indicates one of several alternatives; comparisons treat membership set‑wise.

Error Typing
- Errors are first-class, with a stable code and message (e.g., E_PARSE, E_IO_PERMISSION).
- An error may include position information (line/column/offset) and optional data fields to help diagnose.
- Typed errors allow you to write clear rules (e.g., retry only on specific codes) and to assert behavior in tests.

Contracts (Behavioral Guarantees)
- Contracts define what a node/function promises about input/output and behavior.
- Examples:
  - Input/Output schema: shape and types of event payloads.
  - Ordering and buffering: does a node preserve event order? What is its buffering strategy?
  - Error handling: which error codes may be emitted? Are they retriable?
  - Time and resource use: timeouts, backpressure behavior, or side‑effects (I/O capabilities).

Why this matters
- Clear event and error types improve safety and tooling.
- Contracts make pipelines predictable and testable.
- Lint and test tooling can use these definitions to give better warnings and richer test feedback.

Planned Deliverables (tracked in SPECIFICATION.md)
- Event Schema (events.v1):
  - Define canonical fields (id, timestamp, attempt, trace context).
  - Define payload typing rules and supported container types.
  - JSON representation guidelines for tools and logs.
- Error Schema (diag.v1 alignment):
  - Stable error codes and levels; optional position and data fields.
  - Mapping between runtime errors and lint/test diagnostics.
- Contracts:
  - Node function signatures and I/O shape declarations.
  - Buffering/order guarantees; backpressure policy options and constraints.
  - Capability declarations (e.g., io.* usage only in ingress/egress) and enforcement hooks.
- Validation & Tests:
  - Schema validators for events and errors.
  - Lint checks leveraging contracts (e.g., policy smells) where applicable.
  - Test harness (later phase) to assert contract conformance.

Schema Reference
- Canonical JSON Schema file: `src/schemas/events/schema.json`.
- Tooling can print the embedded schema via the hidden CLI: `ami events schema --print`.

How to get started
- Read the overview above, then open the `.docx` sections 1.7 and 2.2 for the precise definitions.
- When writing pipeline code, document expected input/output shapes and any ordering/buffering assumptions.
- Use `ami lint` to catch capability and pipeline policy issues early; `ami test` can assert parser‑level correctness
today.

Notes
- Until runtime execution tests are integrated, `ami test` focuses on parser‑level assertions. Contract execution
tests will come with the runtime harness phase.

Tooling tip: print the embedded events schema with `ami events schema --print` (hidden command).
