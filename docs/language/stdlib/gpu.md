# AMI stdlib: gpu (stub)

Status: Metal availability + device enumeration on macOS; other operations are stubs.

This package provides a placeholder GPU interface aligned with the work tracker
specification (S-8 gpu package). It intentionally does not perform real GPU
work; all backend operations return deterministic sentinel errors. The goals are:

- Provide stable function names and types for code and docs.
- Enable AMI code to import `gpu` and probe availability.
- Model Owned-like release semantics on opaque handles at the Go layer.

Go-level API (subset):
- Discovery: `gpu.CudaAvailable`, `gpu.MetalAvailable`, `gpu.OpenCLAvailable`.
- CUDA: context, buffers, modules, kernels, and launch (all return `ErrUnavailable`).
- Metal (Darwin): `MetalAvailable()` returns true when a Metal device exists; `MetalDevices()` enumerates devices with names; `MetalCreateContext(Device)` returns an owned `Context`. Library/pipeline/dispatch/buffers remain `ErrUnavailable` for now.
- OpenCL: platform discovery, context, program, kernel, launch, and buffers (all return `ErrUnavailable`).

Blocking semantics and error propagation
- `gpu.Blocking(f func() error) error`: Runs f, blocks until it returns, converts panics to errors.
- `gpu.BlockingSubmit(submit func(done chan<- error)) error`: Creates a completion channel, invokes submit with it, and blocks until an error (possibly nil) is sent.
- Backend helpers: `MetalDispatchBlocking`, `CudaLaunchBlocking`, `OpenCLLaunchBlocking` wrap their respective launch calls with Blocking.

AMI-level stub for blocking
- The compiler’s built-in `gpu` package includes a `BlockingSubmit(any) (Error<any>)` signature as a placeholder that the codegen can lower to a blocking launch wrapper.

AMI-level stub (compiler built-in):
- Signatures for `gpu.CudaAvailable()`, `gpu.MetalAvailable()`, `gpu.OpenCLAvailable()`.

Diagnostics and determinism:
- All unimplemented operations return clear sentinel errors; zero or double releases
  return `ErrInvalidHandle` at the Go layer.

Future work will replace stubs with real backends behind explicit availability checks.
