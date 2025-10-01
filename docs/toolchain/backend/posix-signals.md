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
- Trampolines and helpers in LLVM IR (scaffold):
  - `i64 ami_rt_signal_token_for(i64 sig)` — maps OS `signum` to a handler token via `@ami_signal_handlers[signum % 64]`.
  - `void ami_rt_posix_trampoline(i32 signum)` — reads token, fetches thunk via `ami_rt_get_handler_thunk(token)`, and calls it if non‑null.
  - `void ami_rt_os_signal_enable(i64 sig)`, `void ami_rt_os_signal_disable(i64 sig)` — placeholders for future `sigaction` wiring.
- Real implementation later links libc (`sigaction`, `sigemptyset`, etc.) and sets a single C trampoline per signal:
  1. Translate `SignalType` to native signal numbers (e.g., SIGINT=2, SIGTERM=15).
  2. Install one global handler (trampoline) per signal that:
     - Reads `@ami_signal_handlers[sig]` to get the current `handler_token`.
     - Calls `ami_rt_get_handler_thunk(token)` to fetch `fp`.
     - If `fp != NULL`, tail‑calls `fp()`; otherwise returns promptly.
  3. Provide `ami_rt_posix_install_trampoline(i64 sig)` that wraps `sigaction` registration in platform code and can be invoked from the AMI runtime during initialization or `signal.Register`.
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

Notes
- This is an AMI stdlib package (`package signal`) and AMI runtime backend. Nothing here uses or depends on Go’s `os/signal`.
