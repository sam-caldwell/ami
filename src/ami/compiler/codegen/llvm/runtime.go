package llvm

import (
    "os"
    "path/filepath"
    "strings"
)

// RuntimeLL returns a minimal LLVM IR module string providing runtime symbols
// required by generated code and, optionally, a trivial entrypoint `main`.
// The module sets the provided target triple when non-empty; otherwise uses DefaultTriple.
func RuntimeLL(triple string, withMain bool) string {
    if triple == "" { triple = DefaultTriple }
    // Keep output deterministic and minimal.
    // Provide no-op implementations for a small set of runtime functions used by scaffolding.
    // main returns 0 to allow linking an executable during early bring-up.
    s := "; ModuleID = \"ami:runtime\"\n" +
        "target triple = \"" + triple + "\"\n\n" +
        "; minimal runtime stubs for bring-up\n" +
        "define void @ami_rt_panic(i32 %code) {\n" +
        "entry:\n  ret void\n}\n\n" +
        "define ptr @ami_rt_alloc(i64 %size) {\n" +
        "entry:\n  ret ptr null\n}\n\n"
    // zeroization helper: overwrites n bytes at p with 0x00 deterministically
    s += "define void @ami_rt_zeroize(ptr %p, i64 %n) {\n" +
        "entry:\n  br label %loop\n" +
        "loop:\n  %i = phi i64 [ 0, %entry ], [ %next, %body ]\n  %done = icmp uge i64 %i, %n\n  br i1 %done, label %exit, label %body\n" +
        "body:\n  %addr = getelementptr i8, ptr %p, i64 %i\n  store i8 0, ptr %addr, align 1\n  %next = add i64 %i, 1\n  br label %loop\n" +
        "exit:\n  ret void\n}\n\n"
    // Owned ABI using heap-allocated handle { i8* data; i64 len }
    s += "%Owned = type { i8*, i64 }\n"
    s += "declare ptr @malloc(i64)\n"
    s += "declare void @free(ptr)\n"
    s += "declare void @llvm.memcpy.p0.p0.i64(ptr, ptr, i64, i1)\n\n"

    s += "define ptr @ami_rt_owned_new(i8* %data, i64 %len) {\n" +
        "entry:\n  %mem = call ptr @malloc(i64 16)\n  ; allocate data buffer and copy\n  %buf = call ptr @malloc(i64 %len)\n  call void @llvm.memcpy.p0.p0.i64(ptr %buf, ptr %data, i64 %len, i1 false)\n  ; write handle fields\n  %pfield = bitcast ptr %mem to ptr\n  store ptr %buf, ptr %pfield, align 8\n  %lenptr.i8 = getelementptr i8, ptr %mem, i64 8\n  %lfield = bitcast ptr %lenptr.i8 to ptr\n  store i64 %len, ptr %lfield, align 8\n  ret ptr %mem\n}\n\n"

    s += "define i64 @ami_rt_owned_len(ptr %h) {\n" +
        "entry:\n  %lenptr.i8 = getelementptr i8, ptr %h, i64 8\n  %lfield = bitcast ptr %lenptr.i8 to ptr\n  %l = load i64, ptr %lfield, align 8\n  ret i64 %l\n}\n\n"

    s += "define ptr @ami_rt_owned_ptr(ptr %h) {\n" +
        "entry:\n  %pfield = bitcast ptr %h to ptr\n  %p = load ptr, ptr %pfield, align 8\n  ret ptr %p\n}\n\n"

    // Released-handles guard table (fixed-size)
    s += "@ami_released_tab = private global [256 x ptr] zeroinitializer\n"
    s += "@ami_released_idx = private global i64 0\n\n"
    s += "define void @ami_rt_zeroize_owned(ptr %h) {\n" +
        "entry:\n  ; guard: check if handle already released\n  %idx0 = load i64, ptr @ami_released_idx, align 8\n  %cap = add i64 0, 256\n  %limit = call i64 @llvm.umin.i64(i64 %idx0, i64 %cap)\n  br label %gloop\n" +
        "gloop:\n  %gi = phi i64 [ 0, %entry ], [ %ginext, %gcont ]\n  %gdone = icmp uge i64 %gi, %limit\n  br i1 %gdone, label %gexit, label %gbody\n" +
        "gbody:\n  %gt = getelementptr [256 x ptr], ptr @ami_released_tab, i64 0, i64 %gi\n  %gh = load ptr, ptr %gt, align 8\n  %geq = icmp eq ptr %gh, %h\n  br i1 %geq, label %gfound, label %gcont\n" +
        "gfound:\n  ret void\n" +
        "gcont:\n  %ginext = add i64 %gi, 1\n  br label %gloop\n" +
        "gexit:\n  ; perform zeroize + free, then record handle\n  %p = call ptr @ami_rt_owned_ptr(ptr %h)\n  %n = call i64 @ami_rt_owned_len(ptr %h)\n  call void @ami_rt_zeroize(ptr %p, i64 %n)\n  call void @free(ptr %p)\n  call void @free(ptr %h)\n  %idx1 = load i64, ptr @ami_released_idx, align 8\n  %slot = urem i64 %idx1, %cap\n  %rt = getelementptr [256 x ptr], ptr @ami_released_tab, i64 0, i64 %slot\n  store ptr %h, ptr %rt, align 8\n  %idx2 = add i64 %idx1, 1\n  store i64 %idx2, ptr @ami_released_idx, align 8\n  ret void\n}\n\n"
    s += "declare i64 @llvm.umin.i64(i64, i64)\n\n"
    // Error ABI: heap-allocated handle with code, message pointer and length.
    // Layout: { i32 code; i8* msg; i64 len }
    s += "%Error = type { i32, i8*, i64 }\n\n"
    s += "define ptr @ami_rt_error_new(i32 %code, i8* %msg, i64 %len) {\n" +
        "entry:\n  %mem = call ptr @malloc(i64 24)\n  ; store code\n  %cf = bitcast ptr %mem to ptr\n  store i32 %code, ptr %cf, align 4\n  ; allocate and copy message\n  %buf = call ptr @malloc(i64 %len)\n  call void @llvm.memcpy.p0.p0.i64(ptr %buf, ptr %msg, i64 %len, i1 false)\n  ; store msg ptr\n  %msgptr.i8 = getelementptr i8, ptr %mem, i64 8\n  %mp = bitcast ptr %msgptr.i8 to ptr\n  store ptr %buf, ptr %mp, align 8\n  ; store len\n  %lenptr.i8 = getelementptr i8, ptr %mem, i64 16\n  %lp = bitcast ptr %lenptr.i8 to ptr\n  store i64 %len, ptr %lp, align 8\n  ret ptr %mem\n}\n\n"
    s += "define i32 @ami_rt_error_code(ptr %e) {\n" +
        "entry:\n  %cf = bitcast ptr %e to ptr\n  %c = load i32, ptr %cf, align 4\n  ret i32 %c\n}\n\n"
    s += "define ptr @ami_rt_error_msg(ptr %e) {\n" +
        "entry:\n  %msgptr.i8 = getelementptr i8, ptr %e, i64 8\n  %mp = bitcast ptr %msgptr.i8 to ptr\n  %p = load ptr, ptr %mp, align 8\n  ret ptr %p\n}\n\n"
    s += "define i64 @ami_rt_error_len(ptr %e) {\n" +
        "entry:\n  %lenptr.i8 = getelementptr i8, ptr %e, i64 16\n  %lp = bitcast ptr %lenptr.i8 to ptr\n  %l = load i64, ptr %lp, align 8\n  ret i64 %l\n}\n\n"
    s += "define void @ami_rt_error_free(ptr %e) {\n" +
        "entry:\n  %msgptr.i8 = getelementptr i8, ptr %e, i64 8\n  %mp = bitcast ptr %msgptr.i8 to ptr\n  %p = load ptr, ptr %mp, align 8\n  call void @free(ptr %p)\n  call void @free(ptr %e)\n  ret void\n}\n\n"
    // GPU blocking submit: until an explicit runtime queue is added, accept an opaque
    // arg and return success (null Error). Callers should prefer direct dispatch lowerings.
    s += "define ptr @ami_rt_gpu_blocking_submit(ptr %arg) {\nentry:\n  ret ptr null\n}\n\n"

    // Metal runtime symbols: on Darwin targets, declare and let the Objective-C shim
    // provide real implementations at link time. On non-Darwin targets, emit stub defs.
    if strings.Contains(triple, "darwin") {
        s += "declare i1 @ami_rt_metal_available()\n"
        s += "declare ptr @ami_rt_metal_devices()\n"
        s += "declare ptr @ami_rt_metal_ctx_create(ptr)\n"
        s += "declare void @ami_rt_metal_ctx_destroy(ptr)\n"
        s += "declare ptr @ami_rt_metal_lib_compile(ptr)\n"
        s += "declare ptr @ami_rt_metal_pipe_create(ptr, ptr)\n"
        s += "declare ptr @ami_rt_metal_alloc(i64)\n"
        s += "declare void @ami_rt_metal_free(ptr)\n"
        s += "declare void @ami_rt_metal_copy_to_device(ptr, ptr, i64)\n"
        s += "declare void @ami_rt_metal_copy_from_device(ptr, ptr, i64)\n"
        s += "declare ptr @ami_rt_metal_dispatch_blocking(ptr, ptr, i64, i64, i64, i64, i64, i64)\n\n"
    } else {
        s += "define i1 @ami_rt_metal_available() {\nentry:\n  ret i1 0\n}\n\n"
        s += "define ptr @ami_rt_metal_devices() {\nentry:\n  ret ptr null\n}\n\n"
        s += "define ptr @ami_rt_metal_ctx_create(ptr %dev) {\nentry:\n  ret ptr null\n}\n\n"
        s += "define void @ami_rt_metal_ctx_destroy(ptr %ctx) {\nentry:\n  ret void\n}\n\n"
        s += "define ptr @ami_rt_metal_lib_compile(ptr %src) {\nentry:\n  ret ptr null\n}\n\n"
        s += "define ptr @ami_rt_metal_pipe_create(ptr %lib, ptr %name) {\nentry:\n  ret ptr null\n}\n\n"
        s += "define ptr @ami_rt_metal_alloc(i64 %n) {\nentry:\n  ret ptr null\n}\n\n"
        s += "define void @ami_rt_metal_free(ptr %buf) {\nentry:\n  ret void\n}\n\n"
        s += "define void @ami_rt_metal_copy_to_device(ptr %dst, ptr %src, i64 %n) {\nentry:\n  ret void\n}\n\n"
        s += "define void @ami_rt_metal_copy_from_device(ptr %dst, ptr %src, i64 %n) {\nentry:\n  ret void\n}\n\n"
        s += "define ptr @ami_rt_metal_dispatch_blocking(ptr %ctx, ptr %pipe, i64 %gx, i64 %gy, i64 %gz, i64 %tx, i64 %ty, i64 %tz) {\nentry:\n  ret ptr null\n}\n\n"
    }

    // String/Slice length helpers (scaffold): return 0 until full ABI is wired.
    s += "define i64 @ami_rt_string_len(ptr %s) {\nentry:\n  ret i64 0\n}\n\n"
    s += "define i64 @ami_rt_slice_len(ptr %sl) {\nentry:\n  ret i64 0\n}\n\n"

    // No-op ingress spawner stub; real implementation will create threads/processes per ingress trigger.
    s += "define void @ami_rt_spawn_ingress(ptr %name) {\nentry:\n  ret void\n}\n\n"
    // Signal registration: opaque handler tokens by signal (mod 64)
    s += "@ami_signal_handlers = private global [64 x i64] zeroinitializer\n\n"
    s += "define void @ami_rt_signal_register(i64 %sig, i64 %handler) {\n" +
        "entry:\n  %idx = urem i64 %sig, 64\n  %slot = getelementptr [64 x i64], ptr @ami_signal_handlers, i64 0, i64 %idx\n  store i64 %handler, ptr %slot, align 8\n  ; ensure OS layer is enabled to deliver this signal (stubbed unless POSIX tag)\n  call void @ami_rt_os_signal_enable(i64 %sig)\n  ret void\n}\n\n"

    // Handler thunk registry: token → function pointer (ptr). Not exposed at language ABI boundary.
    s += "@ami_handler_thunks = private global [1024 x ptr] zeroinitializer\n\n"
    s += "define void @ami_rt_install_handler_thunk(i64 %token, ptr %fp) {\n" +
        "entry:\n  %idx = urem i64 %token, 1024\n  %slot = getelementptr [1024 x ptr], ptr @ami_handler_thunks, i64 0, i64 %idx\n  store ptr %fp, ptr %slot, align 8\n  ret void\n}\n\n"
    s += "define ptr @ami_rt_get_handler_thunk(i64 %token) {\n" +
        "entry:\n  %idx = urem i64 %token, 1024\n  %slot = getelementptr [1024 x ptr], ptr @ami_handler_thunks, i64 0, i64 %idx\n  %fp = load ptr, ptr %slot, align 8\n  ret ptr %fp\n}\n\n"
    // OS-specific append (behind build tag)
    s += runtimeOSLL()
    // Math runtime helpers (portable, aggregate returns). Use intrinsics where available.
    s += "\n; math helpers (aggregate returns)\n"
    // sincos: {sin(x), cos(x)}
    s += "define { double, double } @ami_rt_math_sincos(double %x) {\nentry:\n  %s = call double @llvm.sin.f64(double %x)\n  %c = call double @llvm.cos.f64(double %x)\n  %a0 = insertvalue { double, double } undef, double %s, 0\n  %a1 = insertvalue { double, double } %a0, double %c, 1\n  ret { double, double } %a1\n}\n\n"
    // frexp: {frac, exp}
    // Implement via bit ops: this is a simple stub for bring-up; precise semantics can be refined.
    s += "define { double, i64 } @ami_rt_math_frexp(double %x) {\nentry:\n  ; naive stub: return {x, 0} for bring-up\n  %a0 = insertvalue { double, i64 } undef, double %x, 0\n  %a1 = insertvalue { double, i64 } %a0, i64 0, 1\n  ret { double, i64 } %a1\n}\n\n"
    // modf: {intpart, fracpart}
    s += "define { double, double } @ami_rt_math_modf(double %x) {\nentry:\n  %i = fptosi double %x to i64\n  %id = sitofp i64 %i to double\n  %f = fsub double %x, %id\n  %a0 = insertvalue { double, double } undef, double %id, 0\n  %a1 = insertvalue { double, double } %a0, double %f, 1\n  ret { double, double } %a1\n}\n\n"
    // Other helpers
    s += "define double @ami_rt_math_inf(i64 %sign) {\nentry:\n  %isneg = icmp slt i64 %sign, 0\n  %sel = select i1 %isneg, double 0xFFF0000000000000, double 0x7FF0000000000000\n  ret double %sel\n}\n\n"
    s += "define i1 @ami_rt_math_isnan(double %x) {\nentry:\n  %r = fcmp uno double %x, %x\n  ret i1 %r\n}\n\n"
    s += "define i1 @ami_rt_math_isinf(double %x, i64 %sign) {\nentry:\n  %pinf = fcmp oeq double %x, 0x7FF0000000000000\n  %ninf = fcmp oeq double %x, 0xFFF0000000000000\n  %isneg = icmp slt i64 %sign, 0\n  %sel = select i1 %isneg, i1 %ninf, i1 %pinf\n  ret i1 %sel\n}\n\n"
    s += "define i1 @ami_rt_math_signbit(double %x) {\nentry:\n  %neg = fcmp olt double %x, 0.0\n  ret i1 %neg\n}\n\n"

    // NaN constant helper
    s += "define double @ami_rt_math_nan() {\nentry:\n  ret double 0x7FF8000000000000\n}\n\n"

    // Remainder helper (IEEE style approximated via frem for bring-up)
    s += "define double @ami_rt_math_remainder(double %x, double %y) {\nentry:\n  %r = frem double %x, %y\n  ret double %r\n}\n\n"

    // Pow10 helper: compute 10^n for integer n
    s += "define double @ami_rt_math_pow10(i64 %n) {\n" +
        "entry:\n  %isneg = icmp slt i64 %n, 0\n  %absn = select i1 %isneg, i64 sub (i64 0, %n), i64 %n\n  br label %loop\n" +
        "loop:\n  %i = phi i64 [ 0, %entry ], [ %next, %body ]\n  %acc = phi double [ 1.0, %entry ], [ %acc2, %body ]\n  %done = icmp uge i64 %i, %absn\n  br i1 %done, label %exit, label %body\n" +
        "body:\n  %acc2 = fmul double %acc, 10.0\n  %next = add i64 %i, 1\n  br label %loop\n" +
        "exit:\n  %res = select i1 %isneg, double fdiv (double 1.0, %acc), double %acc\n  ret double %res\n}\n\n"
    if withMain {
        s += "define i32 @main() {\nentry:\n  ret i32 0\n}\n"
    }
    return s
}

// WriteRuntimeLL writes the runtime LLVM IR text to the given directory and returns the file path.
func WriteRuntimeLL(dir, triple string, withMain bool) (string, error) {
    if triple == "" { triple = DefaultTriple }
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    path := filepath.Join(dir, "runtime.ll")
    return path, os.WriteFile(path, []byte(RuntimeLL(triple, withMain)), 0o644)
}
