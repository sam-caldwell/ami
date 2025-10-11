Title: AMI Entrypoint Spawn Semantics — Fixed Execution Plan

Objective
- Implement the original spec: the compiler-produced entrypoint (object code) spawns a process per pipeline and a thread per ingress node, with no user-level main(). Keep grammar and language intact.

Non-Goals
- No language changes. No user-level main().
- No behavioral redesign of pipelines, nodes, or grammar.
- No GPU/runtime redesign beyond minimal shims already present.

Deliverables
- Runtime IR/C implementations:
  - ami_rt_spawn_pipeline(ptr name): POSIX process spawn (per pipeline name).
  - ami_rt_spawn_ingress(ptr name): POSIX thread spawn (per ingress name) in the spawned process.
  - Weak stubs on unsupported OSes that compile but return gracefully.
- Entrypoint IR generator:
  - Modify WriteIngressEntrypointLL to:
    1) Deduplicate pipeline names extracted from "pkg.pipeline" identifiers.
    2) For each unique pipeline, emit call to ami_rt_spawn_pipeline("pkg.pipeline").
    3) Emit a call to ami_rt_spawn_ingress("pkg.pipeline") once per ingress edge (or once per ingress node when structure is available).
- Symbol dispatcher:
  - Ensure a deterministic table exists at link time mapping (pkg.pipeline, ingress) -> worker entry function pointer.
  - Reuse the driver’s workers_impl.c/workers_real.c emission where possible to expose symbol names that can be looked up by a C-side dispatcher.
  - Provide a C shim in the runtime build that:
    - Declares extern worker symbols (weak if necessary).
    - Offers lookup helpers used by spawned threads to invoke the right worker entry.

Scope & Sequencing (Do Not Deviate)
1) Entrypoint IR (entry.go):
   - Input: ingress IDs as ["pkg.pipeline", ...].
   - Output main() that:
     a) Calls ami_rt_gpu_probe_init() early (exists).
     b) Builds runtime calls:
        - For each unique pipeline string S, call ami_rt_spawn_pipeline(S).
        - For each ingress occurrence of pipeline S, call ami_rt_spawn_ingress(S).
   - Acceptance: entry.ll text contains the string constants and calls above; tests: entry_write_test extended.

2) POSIX Runtime (new C/IR in runtime.go/OS files):
   - ami_rt_spawn_pipeline(name):
     - If fork()+child path: child process creates ingress threads; parent returns.
     - If posix_spawn: execute same binary with an env var or arg to indicate pipeline S (minimal viable: fork path).
     - Acceptance: returns non-null/0 on success; child path can create threads.
   - ami_rt_spawn_ingress(name):
     - pthread_create() a detached thread that:
       - Resolves worker entry via runtime dispatcher lookup by (pipeline name).
       - Calls the worker stub once (scaffold) to satisfy “thread per ingress” semantics.
     - Acceptance: pthread_create returns 0; thread function calls into worker stub if present.

3) Worker Symbol Dispatcher (C shim, compiled in link stage):
   - Use driver’s existing workers_impl.c/workers_real.c outputs (already emitted in debug) as reference to generate a release shim that declares externs for known workers.
   - Provide functions:
     - ami_rt_find_worker_for_pipeline(ptr name) -> function pointer (or null).
   - Acceptance: link succeeds; no undefined externs; null-safe when worker not present.

4) Build-Link Wiring (build_link.go):
   - Ensure generated entry.o is included when no user-level main() exists (already wired).
   - Add the worker dispatcher C shim object into the env link object list.
   - Acceptance: examples/correct links and binary launches.

5) Tests (minimal, deterministic):
   - Unit: entry_write_test: verify calls to ami_rt_spawn_pipeline/ami_rt_spawn_ingress are emitted for provided ingress IDs.
   - Exec (Linux-only): shared worker that increments a counter or prints; assert that at least one child process and one thread are created per pipeline/ingress.
   - Darwin (compile-only): ensure new symbols compile and link into object files.

Constraints & Notes
- Keep implementation POSIX-first (darwin/linux). Windows paths stubbed.
- Do not parse JSON at runtime to resolve workers. Use compile-time externs.
- Maintain deterministic text/IR emission for stable tests.

Acceptance Criteria (Must All Pass)
- No user-level main in examples/correct.
- entry.ll contains calls to ami_rt_spawn_pipeline and ami_rt_spawn_ingress.
- On POSIX, linking succeeds and running the binary does not crash; threads spawn successfully on darwin/linux with test stubs.
- Exec test proves at least one process per pipeline and one thread per ingress (Linux-only).

Rollback/Contingency
- If pthread/fork causes flakiness on CI, gate the exec test to Linux and keep Darwin compile-only.
- Keep a no-op fallback for platforms without POSIX.

