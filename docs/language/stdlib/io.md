# Stdlib: io (I/O Gating and Wrappers)

The `io` stdlib provides deterministic file and socket wrappers used across AMI. It also enforces runtime I/O capability gating so other stdlib modules (e.g., `trigger`) inherit the same permissions when they use `io`.

Policy
- `Policy{AllowFS, AllowNet}`: global switches for filesystem and network access.
- `SetPolicy(p) / GetPolicy() / ResetPolicy()`: configure or query the current policy.
- `ErrCapabilityDenied`: returned when an operation is blocked by policy.

What’s Gated
- Filesystem: `Open`, `Create`, `OpenFile`, `CreateTemp`, `CreateTempDir`, `Stat`, and all `FHO` methods (read/write/seek/flush/truncate). Standard streams (`Stdin`, `Stdout`, `Stderr`) are treated as file handles and currently fall under FS gating.
- Network: `OpenSocket`, `ConnectSocket`, `ListenSocket`, per‑socket I/O (`Read`, `Write`, `Send`, `SendTo`, `ReadFrom`), deadlines, and server helpers (`Listen`, `Serve`, `ServeContext`). Host/network info (`Hostname`, `Interfaces`) also honors Net gating.

Propagation to Other Modules
- `trigger.net`: builds on `io.Socket`; all network activity is gated via `AllowNet`.
- `trigger.fs`: uses `os.Watch`, which in turn consults `io.Stat` for file state; FS polling respects `AllowFS`.

Granularity
- Current switches are coarse (FS vs Net). If stricter controls are desired (e.g., allow `Stdout` but deny file writes), finer‑grained flags can be added later without breaking callers.

Example
```go
// Deny filesystem; allow network
io.SetPolicy(io.Policy{AllowFS: false, AllowNet: true})
_, err := io.Create("/tmp/x")   // -> ErrCapabilityDenied
_, err = io.ListenUDP("127.0.0.1", 0) // ok
io.ResetPolicy()
```

Notes
- Defaults allow all; the runtime sets policy based on sandbox options before executing a pipeline.
- Errors are synchronous at the callsite to keep behavior deterministic and easy to test.

Ingress/Egress Allowed Ops
- Ingress: `Stdin`, `FileRead`, `DirectoryList`, `FileStat`, `FileSeek`.
- Egress: `Stdout`, `Stderr`, `FileWrite`, `FileCreate`, `FileDelete`, `FileTruncate`, `FileAppend`, `FileChmod`, `FileStat`, `DirectoryCreate`, `DirectoryDelete`, `TempFileCreate`, `TempDirectoryCreate`, `FileRead`, `FileChown`, `FileSeek`.
- Lint rule enforces `io.*` steps appear only as the first (ingress) or last (egress) node, and checks operation family against the allowed set above.

Network
- Ingress: `NetListen` (bind/listen/accept) — allowed at ingress.
- Egress: `NetConnect` (connect/dial) and `NetUdpSend`, `NetTcpSend`, `NetIcmpSend` (send/sendto) — allowed at egress.
- Trust: Under `#pragma trust level=untrusted`, network operations are flagged (`E_TRUST_VIOLATION`).
