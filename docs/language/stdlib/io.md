# Stdlib: io (Files, Stdio, Network)

The `io` module provides deterministic file handles and basic networking utilities, including UDP/TCP sockets, for AMI programs.

API (AMI module `io`)

- Files
  - `func io.open(path string) (File, error)` — open existing file read‑only.
  - `func io.create(path string) (File, error)` — create or truncate for write.
  - `func io.openFile(path string, flag int, perm int) (File, error)` — general open (flags like read/write/append).
  - `method File.read(n int) (bytes, error)`; `method File.write(b bytes) (int, error)`.
  - `method File.seek(offset int64, whence int) (int64, error)`; `method File.pos() (int64, error)`.
  - `method File.len() (int64, error)`; `method File.truncate(n int64) error`; `method File.flush() error`.
  - `method File.name() string`; `method File.close() error` — operations after close return `ErrClosed`.
  - Stdio: `io.stdin`, `io.stdout`, `io.stderr` (special `File` handles; `close()` marks closed but does not close the process stdio).

- Temp/stat & host info
  - `func io.createTemp([dir], [dir,suffix]) (File, error)`; `func io.createTempDir() (string, error)`.
  - `type FileInfo { name string, size int64, mode int, modTime Time, isDir bool }`.
  - `func io.stat(path string) (FileInfo, error)` — file metadata snapshot.
  - `func io.hostname() (string, error)` — system hostname.
  - `type NetInterface { index int, name string, mtu int, flags string, addrs slice<string> }`.
  - `func io.interfaces() (slice<NetInterface>, error)` — list interfaces and addresses.

- Sockets
  - `const io.TCP, io.UDP` (ICMP not implemented).
  - `type Socket` — UDP/TCP abstraction with buffered writes and receive handlers.
  - `func io.openSocket(proto string, addr string, port int) (Socket, error)` —
    - UDP: bind `addr:port` (use `0` for ephemeral); use `write` + `sendTo` to transmit.
    - TCP: connect to `addr:port`; use `write` + `send` to transmit.
  - `func io.listenSocket(proto string, addr string, port int) (Socket, error)` —
    - UDP: same as `openSocket(UDP, ...)` (bound listener).
    - TCP: create a listening server socket on `addr:port` for incoming connections.
  - `method Socket.write(b bytes) (int, error)`; `method Socket.send() error`; `method Socket.sendTo(host string, port int) error` (UDP);
    `method Socket.listen(handler func(bytes)) error`; `method Socket.close() error`.

Notes
- `ErrClosed` is returned by file and socket methods after the handle is closed.
- UDP listeners cannot use `send()`; use `sendTo()` with a remote address.
- TCP `listen(handler)` supports both server sockets (accept loop) and connected client sockets (read loop).

Examples (AMI)
- File round‑trip
  ```
  import io
  var f, _ = io.create("/tmp/demo.txt")
  _, _ = f.write(stringToBytes("hello"))
  _ = f.flush()
  _ = f.close()
  var f2, _ = io.open("/tmp/demo.txt")
  var b, _ = f2.read(8)
  _ = f2.close()
  ```

- UDP send/receive
  ```
  import io
  var port = 9000
  var srv, _ = io.openSocket(io.UDP, "127.0.0.1", port)
  _ = srv.listen(func(msg bytes){ /* handle msg */ })

  var cli, _ = io.openSocket(io.UDP, "127.0.0.1", 0)
  _, _ = cli.write(stringToBytes("ping"))
  _ = cli.sendTo("127.0.0.1", port)
  ```

- TCP client and simple server
  ```
  import io
  var port = 9100
  var srv, _ = io.listenSocket(io.TCP, "127.0.0.1", port)
  _ = srv.listen(func(b bytes){ /* handle incoming */ })

  var c, _ = io.openSocket(io.TCP, "127.0.0.1", port)
  _, _ = c.write(stringToBytes("hello"))
  _ = c.send()
  ```

