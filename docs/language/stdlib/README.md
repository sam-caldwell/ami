# AMI Standard Library

This directory documents the AMI language standard library modules. These modules are imported and called from AMI source code and provide deterministic facilities for processes, files/sockets, time, logging pipelines, and enums.

- os: process runner, system info, and environment
- io: files, stdio wrappers, hostname/interfaces, UDP/TCP sockets
- time: sleep/now/duration arithmetic and a simple ticker
- logger: buffered pipelines with backpressure and JSON redaction
- enum: descriptor‑driven helpers for generated enums

Open a module guide:

- ./os.md
- ./io.md
- ./time.md
- ./logger.md
- ./enum.md
