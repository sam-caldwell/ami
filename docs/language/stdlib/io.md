# Stdlib: io (Files, Stdio, Network)

The `io` stdlib package provides deterministic file handles and basic networking utilities, including UDP/TCP sockets.

API (Go package `io`)

File handles (`FHO`)
- `type FHO`: file/stream handle that ensures operations after `Close()` fail with `ErrClosed`.
- `Open(name string) (*FHO, error)`: open existing file read‑only.
- `Create(name string) (*FHO, error)`: create or truncate for write.
- `OpenFile(name string, flag int, perm os.FileMode) (*FHO, error)`: general open.
- `(*FHO).Read(p []byte) (int, error)`, `(*FHO).Write(p []byte) (int, error)`; aliases `ReadBytes`, `WriteBytes`.
- `(*FHO).Seek(offset int64, whence int) (int64, error)`, `(*FHO).Pos() (int64, error)`.
- `(*FHO).Length() (int64, error)`, `(*FHO).Truncate(n int64) error`, `(*FHO).Flush() error` (fsync).
- `(*FHO).Name() string`: last known path (safe after close).
- `(*FHO).Close() error`: idempotent close; marks handle closed. Further ops return `ErrClosed`.

Stdio wrappers
- `var Stdin, Stdout, Stderr *FHO`: safe wrappers. `Close()` marks closed but does not close process stdio.

Temp and stat helpers
- `CreateTemp([dir], [dir,suffix]) (*FHO, error)`: create temp file under `os.TempDir()` (optionally under subdir and/or with suffix).
- `CreateTempDir() (string, error)`: create unique temp directory under `os.TempDir()`.
- `type FileInfo { Name string; Size int64; Mode os.FileMode; ModTime time.Time; IsDir bool }`.
- `Stat(path string) (FileInfo, error)`: simplified `os.Stat`.

Host/network info
- `Hostname() (string, error)`: system hostname.
- `type NetInterface { Index int; Name string; MTU int; Flags string; Addrs []string }`.
- `Interfaces() ([]NetInterface, error)`: list available interfaces and addresses.

Sockets
- `type NetProtocol = string`: one of `TCP`, `UDP`, `ICMP` (ICMP not implemented).
- `type Socket`: UDP/TCP abstraction with buffered writes and simple receive handlers.
- `OpenSocket(proto NetProtocol, addr string, port uint16) (*Socket, error)`:
  - UDP: bind a datagram socket on `addr:port` (use `0` for ephemeral). Use `Write` + `SendTo` to transmit.
  - TCP: connect to `addr:port`. Use `Write` + `Send` to transmit.
- `ListenSocket(proto NetProtocol, addr string, port uint16) (*Socket, error)`:
  - UDP: same as `OpenSocket(UDP, ...)` (bound listener).
  - TCP: create a listening server socket on `addr:port` for incoming connections.
- `(*Socket).Write(p []byte) (int, error)`: buffer bytes for next send.
- `(*Socket).Send() error`: send buffered bytes (TCP connected sockets). Not supported for UDP listeners.
- `(*Socket).SendTo(host string, port uint16) error`: UDP send to remote for UDP listeners.
- `(*Socket).Listen(handler func([]byte)) error`: register a callback for each received message.
- `(*Socket).Close() error`: idempotent close. Further ops return `ErrClosed`.

Notes
- `ErrClosed` is returned by file and socket methods after the handle is closed.
- UDP listeners cannot use `Send()`; use `SendTo()` with a remote address.
- TCP `Listen(handler)` supports both server sockets (accept loop) and connected client sockets (read loop).

Examples
- File round‑trip
  ```go
  f, _ := io.Create("/tmp/demo.txt")
  _, _ = f.Write([]byte("hello"))
  _ = f.Flush()
  _ = f.Close()
  f2, _ := io.Open("/tmp/demo.txt")
  buf := make([]byte, 8)
  n, _ := f2.Read(buf)
  _ = f2.Close()
  _ = n
  ```
- UDP send/receive
  ```go
  // Listener on fixed port (example)
  const port uint16 = 9000
  srv, _ := io.OpenSocket(io.UDP, "127.0.0.1", port)
  _ = srv.Listen(func(b []byte){ /* handle b */ })

  // Client
  cli, _ := io.OpenSocket(io.UDP, "127.0.0.1", 0)
  _, _ = cli.Write([]byte("ping"))
  _ = cli.SendTo("127.0.0.1", port)
  ```
- TCP client and simple server
  ```go
  // Server on fixed port (example)
  const port uint16 = 9100
  srv, _ := io.ListenSocket(io.TCP, "127.0.0.1", port)
  _ = srv.Listen(func(b []byte){ /* handle incoming */ })

  // Client connect and send
  c, _ := io.OpenSocket(io.TCP, "127.0.0.1", port)
  _, _ = c.Write([]byte("hello"))
  _ = c.Send()
  ```
