POSIX Signals Integration Plan (Design Sketch)

Goals
- Expose a safe AMI stdlib surface for signals: `enum SignalType`, `Register(sig, handler)`.
- Keep ABI opaque (no raw pointers). Use deterministic handler tokens and runtime thunks.
- Support per‑OS implementations behind build tags without changing the language surface.

Current Building Blocks
- Tokens: Deterministic 64‑bit hashes of handler identifiers (e.g., `H`, `m.H`, or fallback `anon@<offset>`).
- Registration: `ami_rt_signal_register(i64 sig, i64 handler_token)` stores the latest token for a signal.
- Thunks: `ami_rt_install_handler_thunk(i64 token, ptr fp)` installs a function pointer thunk into a private table.
- Lookup: `ami_rt_get_handler_thunk(i64 token) -> ptr` returns the thunk pointer (or null if missing).

POSIX Runtime (build tag: `runtime_posix`)
- Add tiny shims in LLVM IR (scaffold):
  - `ami_rt_os_signal_enable(i64 sig)` and `ami_rt_os_signal_disable(i64 sig)`
- Real implementation later links libc (`sigaction`, `sigemptyset`, etc.) and sets a single C trampoline:
  1. Translate `SignalType` to native signal numbers (e.g., SIGINT=2, SIGTERM=15).
  2. Install one global handler (trampoline) per signal that:
     - Reads `@ami_signal_handlers[sig]` to get the current `handler_token`.
     - Calls `ami_rt_get_handler_thunk(token)` to fetch `fp`.
     - If `fp != NULL`, tail‑calls `fp()`; otherwise returns promptly.
- Handler installation: when AMI code loads handler functions (during init), runtime installs their thunks via
  `ami_rt_install_handler_thunk(token, &handler)`. Tokens remain opaque to AMI user code.

Safety and Determinism
- No raw pointers are exposed across the AMI boundary.
- Tokens are deterministic (FNV‑1a); resolution is runtime‑internal.
- Multi‑arch behavior: all OS work hidden behind `runtime_posix` tag; Windows can have an analogous tag.

Testing Approach
- LLVM‑only tests assert externs and call shapes.
- Runtime roundtrip tests (later): compile a small entry calling install→get and verify pointer equality in C/LLVM,
  linked with the runtime. OS behavior remains behind a tag to avoid flakiness.

