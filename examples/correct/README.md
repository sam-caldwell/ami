Correct Example

This example now builds cleanly in verbose JSON mode and links a host binary when a C toolchain (clang) is available. It exercises basic package setup and a minimal pipeline.

To build:
- From this directory: `../../build/ami build --verbose --json` (or `go run ../../src/cmd/ami build --verbose --json`).

For an inline worker demo using event payload field reads, also see:
- examples/inline-body-demo
