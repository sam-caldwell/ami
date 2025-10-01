# AMI Standard Library

This directory documents the AMI runtime standard library packages. These Go packages provide portable facilities used by the AMI language runtime and CLI, and are stable, deterministic wrappers over common OS and I/O primitives.

- os: lightweight process runner and system/environment helpers
- io: deterministic file handles, temp files/dirs, stdio wrappers, hostname/interfaces, UDP/TCP sockets
- time: sleep, timestamps, and a simple ticker
- logger: buffered/batched sinks with backpressure and JSON redaction
- enum: descriptor-driven enums with JSON/Text support

See individual package guides:

- ./os.md
- ./io.md
- ./time.md
- ./logger.md
- ./enum.md

