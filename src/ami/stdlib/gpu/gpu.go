package gpu

import (
    "errors"
    "os"
    "strings"
)

// Sentinel errors used by the stubbed GPU stdlib.
var (
    ErrUnavailable     = errors.New("gpu: backend unavailable")
    ErrUnimplemented   = errors.New("gpu: unimplemented stub")
    ErrInvalidHandle   = errors.New("gpu: invalid handle")
    ErrAlreadyReleased = errors.New("gpu: already released")
)

// Common opaque handles and descriptors. These are intentionally minimal and
// do not expose internal representation. Release methods provide deterministic
// double-free detection and zeroization of metadata.

// Device represents a compute device for CUDA/Metal or an OpenCL device.
type Device struct {
    Backend string // "cuda" | "metal" | "opencl"
    ID      int
    Name    string
}

// Platform represents an OpenCL platform descriptor.
type Platform struct {
    Vendor  string
    Name    string
    Version string
}

// Context represents a GPU execution context.
type Context struct {
    backend string
    valid   bool
    ctxId   int
}

// Release releases the context. Returns ErrInvalidHandle for zero or released.
func (c *Context) Release() error {
    if c == nil || !c.valid {
        return ErrInvalidHandle
    }
    // backend-specific teardown
    if c.backend == "metal" && c.ctxId > 0 {
        metalDestroyContextByID(c.ctxId)
    }
    c.backend = ""
    c.ctxId = 0
    c.valid = false
    return nil
}

// Buffer represents device memory.
type Buffer struct {
    backend string
    n       int
    valid   bool
    bufId   int
}

// Release releases the buffer. Returns ErrInvalidHandle for zero or released.
func (b *Buffer) Release() error {
    if b == nil || !b.valid {
        return ErrInvalidHandle
    }
    if b.backend == "metal" && b.bufId > 0 {
        metalFreeBufferByID(b.bufId)
    }
    b.backend = ""
    b.n = 0
    b.bufId = 0
    b.valid = false
    return nil
}

// Module represents a CUDA module (PTX/Cubin).
type Module struct{
    valid bool
}

func (m *Module) Release() error {
    if m == nil || !m.valid { return ErrInvalidHandle }
    m.valid = false
    return nil
}

// Kernel represents a CUDA kernel or OpenCL kernel.
type Kernel struct{
    valid bool
}

func (k *Kernel) Release() error {
    if k == nil || !k.valid { return ErrInvalidHandle }
    k.valid = false
    return nil
}

// Library represents a Metal library.
type Library struct{
    valid bool
    libId int
}

func (l *Library) Release() error {
    if l == nil || !l.valid { return ErrInvalidHandle }
    if l.libId > 0 { metalReleaseLibrary(l.libId) }
    l.valid = false
    l.libId = 0
    return nil
}

// Pipeline represents a Metal compute pipeline.
type Pipeline struct{
    valid bool
    pipeId int
}

func (p *Pipeline) Release() error {
    if p == nil || !p.valid { return ErrInvalidHandle }
    if p.pipeId > 0 { metalReleasePipeline(p.pipeId) }
    p.valid = false
    p.pipeId = 0
    return nil
}

// Program represents an OpenCL program.
type Program struct{
    valid bool
}

func (p *Program) Release() error {
    if p == nil || !p.valid { return ErrInvalidHandle }
    p.valid = false
    return nil
}

// --- CUDA backend (stubs) ---

// CudaAvailable reports whether the CUDA backend is available.
func CudaAvailable() bool { return envBoolTrue("AMI_GPU_FORCE_CUDA") }

// CudaDevices lists CUDA devices. Always empty in stub.
func CudaDevices() []Device {
    if CudaAvailable() {
        return []Device{{Backend: "cuda", ID: 0, Name: "cuda-ci-0"}}
    }
    return nil
}

// CudaCreateContext creates a CUDA context (stub: unavailable).
func CudaCreateContext(dev Device) (Context, error) {
    if dev.Backend != "cuda" || dev.ID < 0 {
        return Context{}, ErrInvalidHandle
    }
    return Context{}, ErrUnavailable
}

// CudaDestroyContext destroys a CUDA context (stub validation).
func CudaDestroyContext(ctx Context) error {
    if !ctx.valid { return ErrInvalidHandle }
    if ctx.backend != "cuda" { return ErrInvalidHandle }
    return ErrUnavailable
}

// CudaAlloc allocates device memory (stub: unavailable).
func CudaAlloc(n int) (Buffer, error) {
    if n <= 0 { return Buffer{}, ErrInvalidHandle }
    return Buffer{}, ErrUnavailable
}

// CudaFree frees device memory (stub validation/unavailable).
func CudaFree(buf Buffer) error {
    if !buf.valid { return ErrInvalidHandle }
    if buf.backend != "cuda" { return ErrInvalidHandle }
    return ErrUnavailable
}

// CudaMemcpyHtoD copies host->device (stub: unavailable).
func CudaMemcpyHtoD(dst Buffer, src []byte) error {
    if !dst.valid || dst.backend != "cuda" { return ErrInvalidHandle }
    if len(src) == 0 { return ErrInvalidHandle }
    return ErrUnavailable
}

// CudaMemcpyDtoH copies device->host (stub: unavailable).
func CudaMemcpyDtoH(dst []byte, src Buffer) error {
    if !src.valid || src.backend != "cuda" { return ErrInvalidHandle }
    if len(dst) == 0 { return ErrInvalidHandle }
    return ErrUnavailable
}

// CudaLoadModule loads a PTX module (stub: unavailable).
func CudaLoadModule(ptx string) (Module, error) {
    if strings.TrimSpace(ptx) == "" { return Module{}, ErrInvalidHandle }
    return Module{}, ErrUnavailable
}

// CudaGetKernel retrieves a kernel handle (stub: unavailable).
func CudaGetKernel(mod Module, name string) (Kernel, error) {
    if !mod.valid { return Kernel{}, ErrInvalidHandle }
    if strings.TrimSpace(name) == "" { return Kernel{}, ErrInvalidHandle }
    return Kernel{}, ErrUnavailable
}

// CudaLaunchKernel launches a CUDA kernel (stub: unavailable).
func CudaLaunchKernel(ctx Context, k Kernel, grid, block [3]uint32, sharedMem uint32, args ...any) error {
    if !ctx.valid || ctx.backend != "cuda" { return ErrInvalidHandle }
    if !k.valid { return ErrInvalidHandle }
    if grid[0] == 0 || grid[1] == 0 || grid[2] == 0 { return ErrInvalidHandle }
    if block[0] == 0 || block[1] == 0 || block[2] == 0 { return ErrInvalidHandle }
    return ErrUnavailable
}

// CudaLaunchBlocking wraps CudaLaunchKernel with panic-safe blocking semantics.
func CudaLaunchBlocking(ctx Context, k Kernel, grid, block [3]uint32, sharedMem uint32, args ...any) error {
    return Blocking(func() error { return CudaLaunchKernel(ctx, k, grid, block, sharedMem, args...) })
}

// MetalDispatchBlocking wraps MetalDispatch with panic-safe blocking semantics.
func MetalDispatchBlocking(ctx Context, p Pipeline, grid, threadsPerGroup [3]uint32, args ...any) error {
    return Blocking(func() error { return MetalDispatch(ctx, p, grid, threadsPerGroup, args...) })
}

// --- OpenCL backend (stubs) ---

// OpenCLAvailable reports whether the OpenCL backend is available.
func OpenCLAvailable() bool { return envBoolTrue("AMI_GPU_FORCE_OPENCL") }

// OpenCLPlatforms lists OpenCL platforms. Always empty in stub.
func OpenCLPlatforms() []Platform {
    if OpenCLAvailable() {
        return []Platform{{Vendor: "CI", Name: "OpenCL-Dummy", Version: "1.2"}}
    }
    return nil
}

// OpenCLDevices enumerates devices for a given platform (env-forced dummy).
func OpenCLDevices(p Platform) []Device {
    if !OpenCLAvailable() { return nil }
    if p.Name == "" && p.Vendor == "" && p.Version == "" { return nil }
    return []Device{{Backend: "opencl", ID: 0, Name: "opencl-ci-0"}}
}

// OpenCLCreateContext creates an OpenCL context (stub: unavailable).
func OpenCLCreateContext(p Platform) (Context, error) {
    if p.Vendor == "" && p.Name == "" && p.Version == "" {
        return Context{}, ErrInvalidHandle
    }
    return Context{}, ErrUnavailable
}

// OpenCLAlloc allocates device memory (stub: unavailable).
func OpenCLAlloc(n int) (Buffer, error) {
    if n <= 0 { return Buffer{}, ErrInvalidHandle }
    return Buffer{}, ErrUnavailable
}

// OpenCLFree frees device memory (stub validation/unavailable).
func OpenCLFree(buf Buffer) error {
    if !buf.valid { return ErrInvalidHandle }
    if buf.backend != "opencl" { return ErrInvalidHandle }
    return ErrUnavailable
}

// OpenCLBuildProgram builds an OpenCL program (stub: unavailable).
func OpenCLBuildProgram(src string) (Program, error) {
    if strings.TrimSpace(src) == "" { return Program{}, ErrInvalidHandle }
    return Program{}, ErrUnavailable
}

// OpenCLGetKernel retrieves a kernel handle (stub: unavailable).
func OpenCLGetKernel(prog Program, name string) (Kernel, error) {
    if !prog.valid { return Kernel{}, ErrInvalidHandle }
    if strings.TrimSpace(name) == "" { return Kernel{}, ErrInvalidHandle }
    return Kernel{}, ErrUnavailable
}

// OpenCLLaunchKernel launches an OpenCL kernel (stub: unavailable).
func OpenCLLaunchKernel(ctx Context, k Kernel, global, local [3]uint64, args ...any) error {
    if !ctx.valid || ctx.backend != "opencl" { return ErrInvalidHandle }
    if !k.valid { return ErrInvalidHandle }
    if global[0] == 0 || global[1] == 0 || global[2] == 0 { return ErrInvalidHandle }
    if local[0] == 0 || local[1] == 0 || local[2] == 0 { return ErrInvalidHandle }
    return ErrUnavailable
}

// OpenCLLaunchBlocking wraps OpenCLLaunchKernel with panic-safe blocking semantics.
func OpenCLLaunchBlocking(ctx Context, k Kernel, global, local [3]uint64, args ...any) error {
    return Blocking(func() error { return OpenCLLaunchKernel(ctx, k, global, local, args...) })
}

// --- helpers ---

func envBoolTrue(name string) bool {
    v := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
    return v == "1" || v == "true" || v == "yes" || v == "on"
}

// Explain formats a deterministic message for GPU stub errors.
func Explain(backend, op string, err error) string {
    msg := "ok"
    switch err {
    case nil:
        msg = "ok"
    case ErrInvalidHandle:
        msg = "invalid handle"
    case ErrUnavailable:
        msg = "backend unavailable"
    case ErrUnimplemented:
        msg = "unimplemented"
    default:
        msg = err.Error()
    }
    return "gpu/" + backend + " " + op + ": " + msg
}

func CudaExplain(op string, err error) string   { return Explain("cuda", op, err) }
func OpenCLExplain(op string, err error) string { return Explain("opencl", op, err) }
func MetalExplain(op string, err error) string  { return Explain("metal", op, err) }
