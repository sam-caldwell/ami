package llvm

import (
    "os"
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
    s += "declare i64 @llvm.umin.i64(i64, i64)\n"
    s += "declare i64 @llvm.ctlz.i64(i64, i1)\n\n"
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
    if strings.Contains(triple, "apple-") || strings.Contains(triple, "apple-macosx") {
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
        // Ex variant with explicit argument binding arrays
        s += "declare ptr @ami_rt_metal_dispatch_blocking_ex(ptr, ptr, i64, i64, i64, i64, i64, i64, ptr, i64, ptr, ptr, i64*)\n\n"
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
        s += "define ptr @ami_rt_metal_dispatch_blocking_ex(ptr %ctx, ptr %pipe, i64 %gx, i64 %gy, i64 %gz, i64 %tx, i64 %ty, i64 %tz, ptr %kinds, i64 %argc, ptr %bufs, ptr %bytes, i64* %lens) {\nentry:\n  ret ptr null\n}\n\n"
    }

    // String/Slice length helpers (scaffold): return 0 until full ABI is wired.
    s += "define i64 @ami_rt_string_len(ptr %s) {\nentry:\n  ret i64 0\n}\n\n"
    s += "define i64 @ami_rt_slice_len(ptr %sl) {\nentry:\n  ret i64 0\n}\n\n"

    // No-op ingress spawner stub; real implementation will create threads/processes per ingress trigger.
    s += "define void @ami_rt_spawn_ingress(ptr %name) {\nentry:\n  ret void\n}\n\n"

    // GPU probe: record availability of backends in a bitmask global.
    // Bit layout: 0=metal, 1=cuda, 2=opencl
    s += "@ami_rt_gpu_mask = private global i64 0\n\n"
    // getenv only needed when probing env-gated backends
    allowMetal, allowCuda, allowOpenCL := gpuAllowedBackends()
    // Emit string constants and getenv declaration for env-gated probes as needed
    if allowCuda || allowOpenCL {
        s += "declare ptr @getenv(ptr)\n\n"
    }
    // Provide weak default stubs for CUDA/OpenCL availability so host shims can override.
    if allowCuda {
        s += "define weak i1 @ami_rt_cuda_available() {\nentry:\n  ret i1 0\n}\n\n"
    }
    if allowOpenCL {
        s += "define weak i1 @ami_rt_opencl_available() {\nentry:\n  ret i1 0\n}\n\n"
    }
    if allowCuda {
        key := "AMI_GPU_FORCE_CUDA"
        esc := encodeCString(key)
        n := len(key) + 1
        s += "@.gpu.env.cuda = private constant [" + itoa(n) + " x i8] c\"" + esc + "\"\n"
    }
    if allowOpenCL {
        key := "AMI_GPU_FORCE_OPENCL"
        esc := encodeCString(key)
        n := len(key) + 1
        s += "@.gpu.env.opencl = private constant [" + itoa(n) + " x i8] c\"" + esc + "\"\n"
    }
    // Probe function definition
    s += "define void @ami_rt_gpu_probe_init() {\nentry:\n  %mask = alloca i64, align 8\n  store i64 0, ptr %mask, align 8\n"
    // Darwin: query metal via extern if allowed
    if (strings.Contains(triple, "apple-") || strings.Contains(triple, "apple-macosx")) && allowMetal {
        s += "  %m = call i1 @ami_rt_metal_available()\n"
        s += "  %mext = zext i1 %m to i64\n"
        s += "  %mshift = shl i64 %mext, 0\n"
        s += "  %mcur = load i64, ptr %mask, align 8\n"
        s += "  %mnew = or i64 %mcur, %mshift\n"
        s += "  store i64 %mnew, ptr %mask, align 8\n"
    }
    // CUDA: env-gated availability
    if allowCuda {
        s += "  %cstr = getelementptr inbounds [" + itoa(len("AMI_GPU_FORCE_CUDA")+1) + " x i8], ptr @.gpu.env.cuda, i64 0, i64 0\n"
        s += "  %cenv = call ptr @getenv(ptr %cstr)\n"
        s += "  %cnz = icmp ne ptr %cenv, null\n"
        s += "  %cavail = call i1 @ami_rt_cuda_available()\n"
        s += "  %cor = or i1 %cnz, %cavail\n"
        s += "  %cz = zext i1 %cor to i64\n"
        s += "  %cshift = shl i64 %cz, 1\n"
        s += "  %ccur = load i64, ptr %mask, align 8\n"
        s += "  %cnew = or i64 %ccur, %cshift\n"
        s += "  store i64 %cnew, ptr %mask, align 8\n"
    }
    // OpenCL: env-gated availability
    if allowOpenCL {
        s += "  %ostr = getelementptr inbounds [" + itoa(len("AMI_GPU_FORCE_OPENCL")+1) + " x i8], ptr @.gpu.env.opencl, i64 0, i64 0\n"
        s += "  %oenv = call ptr @getenv(ptr %ostr)\n"
        s += "  %onz = icmp ne ptr %oenv, null\n"
        s += "  %oavail = call i1 @ami_rt_opencl_available()\n"
        s += "  %oor = or i1 %onz, %oavail\n"
        s += "  %oz = zext i1 %oor to i64\n"
        s += "  %oshift = shl i64 %oz, 2\n"
        s += "  %ocur = load i64, ptr %mask, align 8\n"
        s += "  %onew = or i64 %ocur, %oshift\n"
        s += "  store i64 %onew, ptr %mask, align 8\n"
    }
    s += "  %final = load i64, ptr %mask, align 8\n"
    s += "  store i64 %final, ptr @ami_rt_gpu_mask, align 8\n"
    s += "  ret void\n}\n\n"

    // Accessor: check if a backend bit is set; lazily probes if mask is zero
    s += "define i1 @ami_rt_gpu_has(i64 %which) {\nentry:\n  %m0 = load i64, ptr @ami_rt_gpu_mask, align 8\n  %m0_is_zero = icmp eq i64 %m0, 0\n  br i1 %m0_is_zero, label %probe, label %cont\nprobe:\n  call void @ami_rt_gpu_probe_init()\n  br label %cont\ncont:\n  %mask = load i64, ptr @ami_rt_gpu_mask, align 8\n  %one = shl i64 1, %which\n  %and = and i64 %mask, %one\n  %zero = icmp eq i64 %and, 0\n  %res = xor i1 %zero, true\n  ret i1 %res\n}\n\n"
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
    // frexp: {frac, exp} with IEEE-754 semantics.
    // For normals: frac in [0.5,1) with sign of x; exp such that x = frac * 2^exp.
    // For zeros: return {x, 0}. For Inf/NaN: return {x, 0}. For subnormals: normalize mantissa.
    s += "define { double, i64 } @ami_rt_math_frexp(double %x) {\n" +
        "entry:\n  %ibits = bitcast double %x to i64\n  %Eraw = lshr i64 %ibits, 52\n  %E = and i64 %Eraw, 2047\n  %M = and i64 %ibits, 4503599627370495\n  %sign = and i64 %ibits, -9223372036854775808\n  ; handle zero: M==0 && E==0 -> {x,0}\n  %E_is_zero = icmp eq i64 %E, 0\n  %M_is_zero = icmp eq i64 %M, 0\n  %is_zero = and i1 %E_is_zero, %M_is_zero\n  br i1 %is_zero, label %zero, label %check_special\nzero:\n  %z0 = insertvalue { double, i64 } undef, double %x, 0\n  %z1 = insertvalue { double, i64 } %z0, i64 0, 1\n  ret { double, i64 } %z1\ncheck_special:\n  ; handle Inf/NaN: E==2047 -> {x,0}\n  %is_special = icmp eq i64 %E, 2047\n  br i1 %is_special, label %special, label %branch_norm\nspecial:\n  %s0 = insertvalue { double, i64 } undef, double %x, 0\n  %s1 = insertvalue { double, i64 } %s0, i64 0, 1\n  ret { double, i64 } %s1\nbranch_norm:\n  ; branch based on normal (E!=0) vs subnormal (E==0 && M!=0)\n  br i1 %E_is_zero, label %subnormal, label %normal\nnormal:\n  ; exp = E - 1022 ; frac = sign | (1022<<52) | M\n  %exp_n = sub i64 %E, 1022\n  %expbits_n = shl i64 1022, 52\n  %em_n = or i64 %expbits_n, %M\n  %fbits_n = or i64 %em_n, %sign\n  %frac_n = bitcast i64 %fbits_n to double\n  %n0 = insertvalue { double, i64 } undef, double %frac_n, 0\n  %n1 = insertvalue { double, i64 } %n0, i64 %exp_n, 1\n  ret { double, i64 } %n1\nsubnormal:\n  ; M != 0 here. Normalize: shift M left so that bit 51 is set; exp = h - 1073 where h = floor(log2(M))\n  %clz = call i64 @llvm.ctlz.i64(i64 %M, i1 false)\n  %h = sub i64 63, %clz\n  %sh = sub i64 51, %h\n  %Mshift = shl i64 %M, %sh\n  %mant = and i64 %Mshift, 4503599627370495\n  %exp_s = sub i64 %h, 1073\n  %expbits_s = shl i64 1022, 52\n  %em_s = or i64 %expbits_s, %mant\n  %fbits_s = or i64 %em_s, %sign\n  %frac_s = bitcast i64 %fbits_s to double\n  %snb0 = insertvalue { double, i64 } undef, double %frac_s, 0\n  %snb1 = insertvalue { double, i64 } %snb0, i64 %exp_s, 1\n  ret { double, i64 } %snb1\n}\n\n"
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
    s += "define double @ami_rt_math_remainder(double %x, double %y) {\n" +
        "entry:\n  ; NaN/Inf/zero checks\n  %x_nan = fcmp uno double %x, %x\n  %y_nan = fcmp uno double %y, %y\n  %y_zero = fcmp oeq double %y, 0.0\n  %ax = call double @llvm.fabs.f64(double %x)\n  %ay = call double @llvm.fabs.f64(double %y)\n  %x_inf = fcmp oeq double %ax, 0x7FF0000000000000\n  %y_inf = fcmp oeq double %ay, 0x7FF0000000000000\n  %bad1 = or i1 %x_nan, %y_nan\n  %bad2 = or i1 %bad1, %y_zero\n  %bad = or i1 %bad2, %x_inf\n  br i1 %bad, label %nan, label %cont\nnan:\n  ret double 0x7FF8000000000000\ncont:\n  br i1 %y_inf, label %retx, label %calc\nretx:\n  ret double %x\ncalc:\n  %q = fdiv double %x, %y\n  %n = call double @llvm.roundeven.f64(double %q)\n  %ny = fmul double %n, %y\n  %r = fsub double %x, %ny\n  ret double %r\n}\n\n"

    // Pow10 helper: compute 10^n for integer n
    s += "define double @ami_rt_math_pow10(i64 %n) {\n" +
        "entry:\n  %isneg = icmp slt i64 %n, 0\n  %neg = sub i64 0, %n\n  %absn = select i1 %isneg, i64 %neg, i64 %n\n  br label %loop\n" +
        "loop:\n  %i = phi i64 [ 0, %entry ], [ %next, %body ]\n  %acc = phi double [ 1.0, %entry ], [ %acc2, %body ]\n  %done = icmp uge i64 %i, %absn\n  br i1 %done, label %exit, label %body\n" +
        "body:\n  %acc2 = fmul double %acc, 10.0\n  %next = add i64 %i, 1\n  br label %loop\n" +
        "exit:\n  %inv = fdiv double 1.0, %acc\n  %res = select i1 %isneg, double %inv, double %acc\n  ret double %res\n}\n\n"
    // Additional math runtime helpers
    s += "define double @ami_rt_math_hypot(double %x, double %y) {\n" +
        "entry:\n  %x2 = fmul double %x, %x\n  %y2 = fmul double %y, %y\n  %sum = fadd double %x2, %y2\n  %r = call double @llvm.sqrt.f64(double %sum)\n  ret double %r\n}\n\n"
    s += "define double @ami_rt_math_asinh(double %x) {\n" +
        "entry:\n  %x2 = fmul double %x, %x\n  %x2p1 = fadd double %x2, 1.0\n  %root = call double @llvm.sqrt.f64(double %x2p1)\n  %sum = fadd double %x, %root\n  %res = call double @llvm.log.f64(double %sum)\n  ret double %res\n}\n\n"
    s += "define double @ami_rt_math_acosh(double %x) {\n" +
        "entry:\n  %xm1 = fsub double %x, 1.0\n  %xp1 = fadd double %x, 1.0\n  %s1 = call double @llvm.sqrt.f64(double %xm1)\n  %s2 = call double @llvm.sqrt.f64(double %xp1)\n  %prod = fmul double %s1, %s2\n  %sum = fadd double %x, %prod\n  %res = call double @llvm.log.f64(double %sum)\n  ret double %res\n}\n\n"
    s += "define double @ami_rt_math_atanh(double %x) {\n" +
        "entry:\n  %num = fadd double 1.0, %x\n  %den = fsub double 1.0, %x\n  %div = fdiv double %num, %den\n  %ln = call double @llvm.log.f64(double %div)\n  %half = fmul double 5.000000e-01, %ln\n  ret double %half\n}\n\n"
    s += "define double @ami_rt_math_cbrt(double %x) {\n" +
        "entry:\n  %res = call double @llvm.pow.f64(double %x, double 0x3FD5555555555555)\n  ret double %res\n}\n\n"
    s += "define double @ami_rt_math_dim(double %x, double %y) {\n" +
        "entry:\n  %sub = fsub double %x, %y\n  %lt0 = fcmp olt double %sub, 0.0\n  %res = select i1 %lt0, double 0.0, double %sub\n  ret double %res\n}\n\n"
    s += "define double @ami_rt_math_logb(double %x) {\n" +
        "entry:\n  %bits = bitcast double %x to i64\n  %abits = and i64 %bits, 9223372036854775807\n  %iszero = icmp eq i64 %abits, 0\n  br i1 %iszero, label %ret_zero, label %checkinf\n" +
        "ret_zero:\n  ret double 0xFFF0000000000000\n" +
        "checkinf:\n  %expfield = lshr i64 %bits, 52\n  %exp = and i64 %expfield, 2047\n  %mant = and i64 %bits, 4503599627370495\n  %isexpmax = icmp eq i64 %exp, 2047\n  br i1 %isexpmax, label %inf_or_nan, label %normal_or_sub\n" +
        "inf_or_nan:\n  %ismantzero = icmp eq i64 %mant, 0\n  br i1 %ismantzero, label %ret_pos_inf, label %ret_nan\n" +
        "ret_pos_inf:\n  ret double 0x7FF0000000000000\n" +
        "ret_nan:\n  ret double 0x7FF8000000000000\n" +
        "normal_or_sub:\n  %isnorm = icmp ne i64 %exp, 0\n  br i1 %isnorm, label %ret_norm, label %subnorm\n" +
        "ret_norm:\n  %unbiased = sub i64 %exp, 1023\n  %d = sitofp i64 %unbiased to double\n  ret double %d\n" +
        "subnorm:\n  %frac = and i64 %bits, 4503599627370495\n  %sh = shl i64 %frac, 12\n  %lz = call i64 @llvm.ctlz.i64(i64 %sh, i1 false)\n  %tmp = add i64 %lz, 1022\n  %unbiased2 = sub i64 0, %tmp\n  %d2 = sitofp i64 %unbiased2 to double\n  ret double %d2\n}\n\n"
    s += "define i64 @ami_rt_math_ilogb(double %x) {\n" +
        "entry:\n  %bits = bitcast double %x to i64\n  %abits = and i64 %bits, 9223372036854775807\n  %iszero = icmp eq i64 %abits, 0\n  br i1 %iszero, label %ret_min, label %checkinf2\n" +
        "ret_min:\n  ret i64 -9223372036854775808\n" +
        "checkinf2:\n  %expfield = lshr i64 %bits, 52\n  %exp = and i64 %expfield, 2047\n  %mant = and i64 %bits, 4503599627370495\n  %isexpmax = icmp eq i64 %exp, 2047\n  br i1 %isexpmax, label %ret_max, label %normal_or_sub2\n" +
        "ret_max:\n  ret i64 9223372036854775807\n" +
        "normal_or_sub2:\n  %isnorm = icmp ne i64 %exp, 0\n  br i1 %isnorm, label %ret_norm2, label %subnorm2\n" +
        "ret_norm2:\n  %unbiased = sub i64 %exp, 1023\n  ret i64 %unbiased\n" +
        "subnorm2:\n  %frac = and i64 %bits, 4503599627370495\n  %sh = shl i64 %frac, 12\n  %lz = call i64 @llvm.ctlz.i64(i64 %sh, i1 false)\n  %tmp = add i64 %lz, 1022\n  %unbiased2 = sub i64 0, %tmp\n  ret i64 %unbiased2\n}\n\n"
    if withMain {
        s += "define i32 @main() {\nentry:\n  ret i32 0\n}\n"
    }
    // JSON bridge helpers (stubs): in early bring-up, convert Event handles
    // to a minimal JSON and return not-yet-implemented for payloads.
    // Constants
    evJSON := "{\"schema\":\"events.v1\"}"
    evEsc := encodeCString(evJSON)
    evN := len(evJSON) + 1
    s += "@.json.event.empty = private constant [" + itoa(evN) + " x i8] c\"" + evEsc + "\"\n"
    nulJSON := "null"
    nulEsc := encodeCString(nulJSON)
    nulN := len(nulJSON) + 1
    s += "@.json.null = private constant [" + itoa(nulN) + " x i8] c\"" + nulEsc + "\"\n\n"
    // define ptr @ami_rt_json_to_event(ptr in, i32 inlen)
    s += "define ptr @ami_rt_json_to_event(ptr %in, i32 %inlen) {\nentry:\n  ret ptr null\n}\n\n"
    // define ptr @ami_rt_event_to_json(ptr ev, i32* outlen)
    s += "define ptr @ami_rt_event_to_json(ptr %ev, i32* %outlen) {\nentry:\n  %src = getelementptr inbounds [" + itoa(evN) + " x i8], ptr @.json.event.empty, i64 0, i64 0\n  %len = zext i32 " + itoa(evN) + " to i64\n  %buf = call ptr @malloc(i64 %len)\n  call void @llvm.memcpy.p0.p0.i64(ptr %buf, ptr %src, i64 %len, i1 false)\n  store i32 " + itoa(len(evJSON)) + ", ptr %outlen, align 4\n  ret ptr %buf\n}\n\n"
    // define ptr @ami_rt_payload_to_json(ptr p, i32* outlen)
    s += "define ptr @ami_rt_payload_to_json(ptr %p, i32* %outlen) {\nentry:\n  %src = getelementptr inbounds [" + itoa(nulN) + " x i8], ptr @.json.null, i64 0, i64 0\n  %len = zext i32 " + itoa(nulN) + " to i64\n  %buf = call ptr @malloc(i64 %len)\n  call void @llvm.memcpy.p0.p0.i64(ptr %buf, ptr %src, i64 %len, i1 false)\n  store i32 " + itoa(len(nulJSON)) + ", ptr %outlen, align 4\n  ret ptr %buf\n}\n\n"
    return s
}

// WriteRuntimeLL writes the runtime LLVM IR text to the given directory and returns the file path.
// file writing moved to runtime_write.go to satisfy single-declaration rule

// gpuAllowedBackends inspects the AMI_GPU_BACKENDS env var and returns booleans for
// allowing metal, cuda, opencl in the generated runtime IR. Empty or missing means all allowed.
func gpuAllowedBackends() (metal, cuda, opencl bool) {
    v := os.Getenv("AMI_GPU_BACKENDS")
    if v == "" {
        return true, true, true
    }
    var m, c, o bool
    parts := strings.Split(v, ",")
    for _, p := range parts {
        p = strings.TrimSpace(strings.ToLower(p))
        switch p {
        case "metal":
            m = true
        case "cuda":
            c = true
        case "opencl", "ocl":
            o = true
        }
    }
    return m, c, o
}

// itoa is a tiny integer to string formatter for small positive values used in IR lengths.
// itoa helper provided in emit_util.go
