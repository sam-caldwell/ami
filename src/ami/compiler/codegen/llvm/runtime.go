package llvm

import (
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
    // libc snprintf for primitive number formatting in payload->JSON helpers
    s += "declare i32 @snprintf(ptr, i64, ptr, ...)\n\n"

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
    // Convert error handle to C string (malloc'd) and set outlen
    s += "define ptr @ami_rt_error_to_cstring(ptr %e, i32* %outlen) {\nentry:\n  %msg = call ptr @ami_rt_error_msg(ptr %e)\n  %len64 = call i64 @ami_rt_error_len(ptr %e)\n  %len = trunc i64 %len64 to i32\n  %sz = zext i32 %len to i64\n  %buf = call ptr @malloc(i64 %sz)\n  call void @llvm.memcpy.p0.p0.i64(ptr %buf, ptr %msg, i64 %sz, i1 false)\n  store i32 %len, ptr %outlen, align 4\n  ret ptr %buf\n}\n\n"
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
    // JSON bridge helpers: Introduce a minimal Event handle layout and convert to/from JSON
    // Event handle layout: { i8* payload_ptr; i64 payload_len }
    s += "%Event = type { i8*, i64 }\n\n"
    // Constants for JSON composition
    // Note: The [N x i8] length must match the number of bytes (not the escaped
    // source length). We compute N using the original literal length + 1 for NUL,
    // then escape for the IR text. Copy lengths exclude the trailing NUL.
    prefix := "{\"schema\":\"events.v1\",\"payload\":"
    suffix := "}"
    evEsc := encodeCString(prefix)
    evBytes := len(prefix) + 1 // bytes including NUL
    s += "@.json.event.prefix = private constant [" + itoa(evBytes) + " x i8] c\"" + evEsc + "\"\n"
    sufEsc := encodeCString(suffix)
    sufBytes := len(suffix) + 1 // bytes including NUL
    s += "@.json.event.suffix = private constant [" + itoa(sufBytes) + " x i8] c\"" + sufEsc + "\"\n"
    nulJSON := "null"
    nulEsc := encodeCString(nulJSON)
    nulN := len(nulJSON) + 1
    s += "@.json.null = private constant [" + itoa(nulN) + " x i8] c\"" + nulEsc + "\"\n\n"
    // define ptr @ami_rt_json_to_event(ptr in, i32 inlen)
    s += "define ptr @ami_rt_json_to_event(ptr %in, i32 %inlen) {\n" +
        "entry:\n  %sz = zext i32 %inlen to i64\n  %buf = call ptr @malloc(i64 %sz)\n  call void @llvm.memcpy.p0.p0.i64(ptr %buf, ptr %in, i64 %sz, i1 false)\n  %eh = call ptr @malloc(i64 16)\n  %pfield = bitcast ptr %eh to ptr\n  store ptr %buf, ptr %pfield, align 8\n  %lenptr.i8 = getelementptr i8, ptr %eh, i64 8\n  %lfield = bitcast ptr %lenptr.i8 to ptr\n  store i64 %sz, ptr %lfield, align 8\n  ret ptr %eh\n}\n\n"
    // define ptr @ami_rt_event_to_json(ptr ev, i32* outlen)
    s += "define ptr @ami_rt_event_to_json(ptr %ev, i32* %outlen) {\nentry:\n  ; load payload ptr and len from Event handle\n  %pfield = bitcast ptr %ev to ptr\n  %pp = load ptr, ptr %pfield, align 8\n  %lenptr.i8 = getelementptr i8, ptr %ev, i64 8\n  %lfield = bitcast ptr %lenptr.i8 to ptr\n  %plen = load i64, ptr %lfield, align 8\n  ; compute total size = prefix + payload + suffix (no NUL terminator)\n  %pref = zext i32 " + itoa(evBytes-1) + " to i64\n  %suf = zext i32 " + itoa(sufBytes-1) + " to i64\n  %psum = add i64 %pref, %plen\n  %tot = add i64 %psum, %suf\n  %buf = call ptr @malloc(i64 %tot)\n  ; copy prefix\n  %pfx = getelementptr inbounds [" + itoa(evBytes) + " x i8], ptr @.json.event.prefix, i64 0, i64 0\n  call void @llvm.memcpy.p0.p0.i64(ptr %buf, ptr %pfx, i64 %pref, i1 false)\n  ; copy payload just after prefix\n  %dst1 = getelementptr i8, ptr %buf, i64 %pref\n  call void @llvm.memcpy.p0.p0.i64(ptr %dst1, ptr %pp, i64 %plen, i1 false)\n  ; copy suffix at end\n  %dst2 = getelementptr i8, ptr %buf, i64 %psum\n  %sfx = getelementptr inbounds [" + itoa(sufBytes) + " x i8], ptr @.json.event.suffix, i64 0, i64 0\n  call void @llvm.memcpy.p0.p0.i64(ptr %dst2, ptr %sfx, i64 %suf, i1 false)\n  ; outlen = total length (i32)\n  %tot32 = trunc i64 %tot to i32\n  store i32 %tot32, ptr %outlen, align 4\n  ret ptr %buf\n}\n\n"
    // define ptr @ami_rt_payload_to_json(ptr p, i32* outlen)
    s += "define ptr @ami_rt_payload_to_json(ptr %p, i32* %outlen) {\nentry:\n  %src = getelementptr inbounds [" + itoa(nulN) + " x i8], ptr @.json.null, i64 0, i64 0\n  %len = zext i32 " + itoa(nulN) + " to i64\n  %buf = call ptr @malloc(i64 %len)\n  call void @llvm.memcpy.p0.p0.i64(ptr %buf, ptr %src, i64 %len, i1 false)\n  store i32 " + itoa(len(nulJSON)) + ", ptr %outlen, align 4\n  ret ptr %buf\n}\n\n"

    // Primitive payload JSON helpers
    // Booleans
    s += "@.json.true = private constant [5 x i8] c\"true\\00\"\n"
    s += "@.json.false = private constant [6 x i8] c\"false\\00\"\n"
    s += "define ptr @ami_rt_bool_to_json(i1 %v, i32* %outlen) {\n"
    s += "entry:\n  br i1 %v, label %ltrue, label %lfalse\n"
    s += "ltrue:\n  %src_t = getelementptr inbounds [5 x i8], ptr @.json.true, i64 0, i64 0\n  %len_t = zext i32 4 to i64\n  %buf_t = call ptr @malloc(i64 %len_t)\n  call void @llvm.memcpy.p0.p0.i64(ptr %buf_t, ptr %src_t, i64 %len_t, i1 false)\n  store i32 4, ptr %outlen, align 4\n  ret ptr %buf_t\n"
    s += "lfalse:\n  %src_f = getelementptr inbounds [6 x i8], ptr @.json.false, i64 0, i64 0\n  %len_f = zext i32 5 to i64\n  %buf_f = call ptr @malloc(i64 %len_f)\n  call void @llvm.memcpy.p0.p0.i64(ptr %buf_f, ptr %src_f, i64 %len_f, i1 false)\n  store i32 5, ptr %outlen, align 4\n  ret ptr %buf_f\n}\n\n"

    // Integer formatting using snprintf("%lld") into a heap buffer
    s += "@.fmt.i64 = private constant [5 x i8] c\"%lld\\00\"\n"
    s += "define ptr @ami_rt_i64_to_json(i64 %n, i32* %outlen) {\n"
    s += "entry:\n  ; allocate a temporary buffer; snprintf returns length excluding NUL\n  %buf = call ptr @malloc(i64 64)\n  %fmt = getelementptr inbounds [5 x i8], ptr @.fmt.i64, i64 0, i64 0\n  %written = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %buf, i64 64, ptr %fmt, i64 %n)\n  store i32 %written, ptr %outlen, align 4\n  ret ptr %buf\n}\n\n"

    // Double formatting using snprintf("%.17g")
    // "%.17g" is 5 chars + NUL terminator = 6 bytes
    s += "@.fmt.f64 = private constant [6 x i8] c\"%.17g\\00\"\n"
    s += "define ptr @ami_rt_double_to_json(double %x, i32* %outlen) {\n"
    s += "entry:\n  %buf = call ptr @malloc(i64 64)\n  %fmt = getelementptr inbounds [6 x i8], ptr @.fmt.f64, i64 0, i64 0\n  %written = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %buf, i64 64, ptr %fmt, double %x)\n  store i32 %written, ptr %outlen, align 4\n  ret ptr %buf\n}\n\n"

    // String to JSON with proper escaping per RFC 8259: \" \\ control chars -> \b\f\n\r\t; others <0x20 -> \u00XX
    s += "define ptr @ami_rt_string_to_json(ptr %p, i32* %outlen) {\n"
    s += "entry:\n  %len64 = call i64 @ami_rt_owned_len(ptr %p)\n  %src = call ptr @ami_rt_owned_ptr(ptr %p)\n  ; max size = 6*len + 2 (quotes)\n  %two = shl i64 %len64, 1\n  %four = shl i64 %len64, 2\n  %six = add i64 %two, %four\n  %max = add i64 %six, 2\n  %buf = call ptr @malloc(i64 %max)\n  ; opening quote at pos 0\n  store i8 34, ptr %buf\n  br label %loop\n"
    s += "loop:\n  %i = phi i64 [ 0, %entry ], [ %inext, %cont ]\n  %j = phi i64 [ 1, %entry ], [ %jphi, %cont ]\n  %done = icmp uge i64 %i, %len64\n  br i1 %done, label %after, label %body\n"
    s += "body:\n  %ip = getelementptr i8, ptr %src, i64 %i\n  %b = load i8, ptr %ip, align 1\n  %isq = icmp eq i8 %b, 34\n  %isb = icmp eq i8 %b, 92\n  %escq = or i1 %isq, %isb\n  br i1 %escq, label %esc_qbs, label %check_b\n"
    s += "esc_qbs:\n  %dq1 = getelementptr i8, ptr %buf, i64 %j\n  store i8 92, ptr %dq1\n  %j1 = add i64 %j, 1\n  %dq2 = getelementptr i8, ptr %buf, i64 %j1\n  store i8 %b, ptr %dq2\n  %jout1 = add i64 %j1, 1\n  br label %cont\n"
    s += "check_b:\n  %is_b = icmp eq i8 %b, 8\n  br i1 %is_b, label %esc_b, label %check_f\n"
    s += "esc_b:\n  %db1 = getelementptr i8, ptr %buf, i64 %j\n  store i8 92, ptr %db1\n  %jb1 = add i64 %j, 1\n  %db2 = getelementptr i8, ptr %buf, i64 %jb1\n  store i8 98, ptr %db2\n  %joutb = add i64 %jb1, 1\n  br label %cont\n"
    s += "check_f:\n  %is_f = icmp eq i8 %b, 12\n  br i1 %is_f, label %esc_f, label %check_n\n"
    s += "esc_f:\n  %df1 = getelementptr i8, ptr %buf, i64 %j\n  store i8 92, ptr %df1\n  %jf1 = add i64 %j, 1\n  %df2 = getelementptr i8, ptr %buf, i64 %jf1\n  store i8 102, ptr %df2\n  %joutf = add i64 %jf1, 1\n  br label %cont\n"
    s += "check_n:\n  %is_n = icmp eq i8 %b, 10\n  br i1 %is_n, label %esc_n, label %check_r\n"
    s += "esc_n:\n  %dn1 = getelementptr i8, ptr %buf, i64 %j\n  store i8 92, ptr %dn1\n  %jn1 = add i64 %j, 1\n  %dn2 = getelementptr i8, ptr %buf, i64 %jn1\n  store i8 110, ptr %dn2\n  %joutn = add i64 %jn1, 1\n  br label %cont\n"
    s += "check_r:\n  %is_r = icmp eq i8 %b, 13\n  br i1 %is_r, label %esc_r, label %check_t\n"
    s += "esc_r:\n  %dr1 = getelementptr i8, ptr %buf, i64 %j\n  store i8 92, ptr %dr1\n  %jr1 = add i64 %j, 1\n  %dr2 = getelementptr i8, ptr %buf, i64 %jr1\n  store i8 114, ptr %dr2\n  %joutr = add i64 %jr1, 1\n  br label %cont\n"
    s += "check_t:\n  %is_t = icmp eq i8 %b, 9\n  br i1 %is_t, label %esc_t, label %check_lt\n"
    s += "esc_t:\n  %dt1 = getelementptr i8, ptr %buf, i64 %j\n  store i8 92, ptr %dt1\n  %jt1 = add i64 %j, 1\n  %dt2 = getelementptr i8, ptr %buf, i64 %jt1\n  store i8 116, ptr %dt2\n  %joutt = add i64 %jt1, 1\n  br label %cont\n"
    s += "check_lt:\n  %lt = icmp ult i8 %b, 32\n  br i1 %lt, label %esc_u, label %copy\n"
    s += "esc_u:\n  %b32 = zext i8 %b to i32\n  %hib = lshr i32 %b32, 4\n  %hi = and i32 %hib, 15\n  %lo = and i32 %b32, 15\n  %hi_lt10 = icmp ult i32 %hi, 10\n  %hi_base = select i1 %hi_lt10, i32 48, i32 55\n  %hi_ascii32 = add i32 %hi_base, %hi\n  %hi_ascii = trunc i32 %hi_ascii32 to i8\n  %lo_lt10 = icmp ult i32 %lo, 10\n  %lo_base = select i1 %lo_lt10, i32 48, i32 55\n  %lo_ascii32 = add i32 %lo_base, %lo\n  %lo_ascii = trunc i32 %lo_ascii32 to i8\n  %du0 = getelementptr i8, ptr %buf, i64 %j\n  store i8 92, ptr %du0\n  %ju1 = add i64 %j, 1\n  %du1 = getelementptr i8, ptr %buf, i64 %ju1\n  store i8 117, ptr %du1\n  %ju2 = add i64 %ju1, 1\n  %du2 = getelementptr i8, ptr %buf, i64 %ju2\n  store i8 48, ptr %du2\n  %ju3 = add i64 %ju2, 1\n  %du3 = getelementptr i8, ptr %buf, i64 %ju3\n  store i8 48, ptr %du3\n  %ju4 = add i64 %ju3, 1\n  %du4 = getelementptr i8, ptr %buf, i64 %ju4\n  store i8 %hi_ascii, ptr %du4\n  %ju5 = add i64 %ju4, 1\n  %du5 = getelementptr i8, ptr %buf, i64 %ju5\n  store i8 %lo_ascii, ptr %du5\n  %joutu = add i64 %ju5, 1\n  br label %cont\n"
    s += "copy:\n  %dc = getelementptr i8, ptr %buf, i64 %j\n  store i8 %b, ptr %dc\n  %jnorm = add i64 %j, 1\n  br label %cont\n"
    s += "cont:\n  %inext = add i64 %i, 1\n  %jphi = phi i64 [ %jout1, %esc_qbs ], [ %joutb, %esc_b ], [ %joutf, %esc_f ], [ %joutn, %esc_n ], [ %joutr, %esc_r ], [ %joutt, %esc_t ], [ %joutu, %esc_u ], [ %jnorm, %copy ]\n  br label %loop\n"
    s += "after:\n  %endp = getelementptr i8, ptr %buf, i64 %j\n  store i8 34, ptr %endp\n  %jfin = add i64 %j, 1\n  %olen = trunc i64 %jfin to i32\n  store i32 %olen, ptr %outlen\n  ret ptr %buf\n}\n\n"
    // Structured passthrough: treat %p as Owned handle containing pre-serialized JSON bytes
    s += "define ptr @ami_rt_structured_to_json(ptr %p, i32* %outlen) {\n"
    s += "entry:\n  %len64 = call i64 @ami_rt_owned_len(ptr %p)\n  %src = call ptr @ami_rt_owned_ptr(ptr %p)\n  %buf = call ptr @malloc(i64 %len64)\n  call void @llvm.memcpy.p0.p0.i64(ptr %buf, ptr %src, i64 %len64, i1 false)\n  %len32 = trunc i64 %len64 to i32\n  store i32 %len32, ptr %outlen, align 4\n  ret ptr %buf\n}\n\n"
    // Struct walker: 'S' + i32 count + entries (u8 nameLen, name bytes, u8 kind). Values: array of ptrs
    s += "define ptr @ami_rt_struct_to_json(ptr %td, i64 %vh, i32* %outlen) {\n"
    s += "entry:\n  %vals = inttoptr i64 %vh to ptr\n  %cntp.i8 = getelementptr i8, ptr %td, i64 1\n  %cntp = bitcast ptr %cntp.i8 to ptr\n  %cnt = load i32, ptr %cntp, align 1\n  %n64 = zext i32 %cnt to i64\n  %pos = getelementptr i8, ptr %cntp.i8, i64 4\n  %has = icmp ugt i32 %cnt, 0\n  %cm1 = add i32 %cnt, 4294967295\n  %comma32 = select i1 %has, i32 %cm1, i32 0\n  %comma64 = zext i32 %comma32 to i64\n  %sum0 = add i64 2, %comma64\n  %tmpout = alloca i32, align 4\n  br label %lenloop\n"
    s += "lenloop:\n  %i = phi i64 [ 0, %entry ], [ %inext, %lencont ]\n  %pcur = phi ptr [ %pos, %entry ], [ %pafter, %lencont ]\n  %sum = phi i64 [ %sum0, %entry ], [ %sum2, %lencont ]\n  %done = icmp uge i64 %i, %n64\n  br i1 %done, label %alloc, label %lenbody\n"
    s += "lenbody:\n  %nl = load i8, ptr %pcur, align 1\n  %nl64 = zext i8 %nl to i64\n  %pname = getelementptr i8, ptr %pcur, i64 1\n  %pnend = getelementptr i8, ptr %pname, i64 %nl64\n  %kind = load i8, ptr %pnend, align 1\n  %pafter = getelementptr i8, ptr %pnend, i64 1\n  %vpp = getelementptr ptr, ptr %vals, i64 %i\n  %vhp = load ptr, ptr %vpp, align 8\n  ; branch by kind to compute value length\n  %isk_i = icmp eq i8 %kind, 105\n  %isk_d = icmp eq i8 %kind, 100\n  %isk_b = icmp eq i8 %kind, 98\n  %isk_s = icmp eq i8 %kind, 115\n  br i1 %isk_i, label %len_i, label %chk_d\n"
    s += "len_i:\n  %ival = load i64, ptr %vhp, align 8\n  %tbi = call ptr @ami_rt_i64_to_json(i64 %ival, i32* %tmpout)\n  %vlen32i = load i32, ptr %tmpout, align 4\n  %vleni = zext i32 %vlen32i to i64\n  call void @free(ptr %tbi)\n  br label %len_accum\n"
    s += "chk_d:\n  br i1 %isk_d, label %len_d, label %chk_b\n"
    s += "len_d:\n  %dval = load double, ptr %vhp, align 8\n  %tbd = call ptr @ami_rt_double_to_json(double %dval, i32* %tmpout)\n  %vlen32d = load i32, ptr %tmpout, align 4\n  %vlend = zext i32 %vlen32d to i64\n  call void @free(ptr %tbd)\n  br label %len_accum\n"
    s += "chk_b:\n  br i1 %isk_b, label %len_b, label %chk_s\n"
    s += "len_b:\n  %b8 = load i8, ptr %vhp, align 1\n  %b1 = icmp ne i8 %b8, 0\n  %tbb = call ptr @ami_rt_bool_to_json(i1 %b1, i32* %tmpout)\n  %vlen32b = load i32, ptr %tmpout, align 4\n  %vlenb = zext i32 %vlen32b to i64\n  call void @free(ptr %tbb)\n  br label %len_accum\n"
    s += "chk_s:\n  br i1 %isk_s, label %len_s, label %len_o\n"
    s += "len_s:\n  %tbs = call ptr @ami_rt_string_to_json(ptr %vhp, i32* %tmpout)\n  %vlen32s = load i32, ptr %tmpout, align 4\n  %vlens = zext i32 %vlen32s to i64\n  call void @free(ptr %tbs)\n  br label %len_accum\n"
    s += "len_o:\n  %vleno = call i64 @ami_rt_owned_len(ptr %vhp)\n  br label %len_accum\n"
    s += "len_accum:\n  %vlen = phi i64 [ %vleni, %len_i ], [ %vlend, %len_d ], [ %vlenb, %len_b ], [ %vlens, %len_s ], [ %vleno, %len_o ]\n  br label %len_accum2\n"
    s += "len_accum2:\n  %tmp1 = add i64 %sum, 2\n  %tmp2 = add i64 %tmp1, %nl64\n  %tmp3 = add i64 %tmp2, 1\n  %sum2 = add i64 %tmp3, %vlen\n  %inext = add i64 %i, 1\n  br label %lencont\n"
    s += "lencont:\n  br label %lenloop\n"
    s += "alloc:\n  %buf = call ptr @malloc(i64 %sum)\n  store i8 123, ptr %buf\n  %outp = getelementptr i8, ptr %buf, i64 1\n  br label %emitloop\n"
    s += "emitloop:\n  %j = phi i64 [ 0, %alloc ], [ %jnext, %emitcont ]\n  %pcur2 = phi ptr [ %pos, %alloc ], [ %pafter2, %emitcont ]\n  %dst = phi ptr [ %outp, %alloc ], [ %nextdst, %emitcont ]\n  %done2 = icmp uge i64 %j, %n64\n  br i1 %done2, label %after, label %emitbody\n"
    s += "emitbody:\n  %gt0 = icmp ugt i64 %j, 0\n  br i1 %gt0, label %comma, label %noc\n"
    s += "comma:\n  store i8 44, ptr %dst\n  %dstc_c = getelementptr i8, ptr %dst, i64 1\n  br label %nocont\n"
    s += "noc:\n  %dstc_n = getelementptr i8, ptr %dst, i64 0\n  br label %nocont\n"
    s += "nocont:\n  %dstc = phi ptr [ %dstc_c, %comma ], [ %dstc_n, %noc ]\n  %nl2 = load i8, ptr %pcur2, align 1\n  %nl64b = zext i8 %nl2 to i64\n  %pname2 = getelementptr i8, ptr %pcur2, i64 1\n  %pnend2 = getelementptr i8, ptr %pname2, i64 %nl64b\n  %kind2 = load i8, ptr %pnend2, align 1\n  %pafter2 = getelementptr i8, ptr %pnend2, i64 1\n  store i8 34, ptr %dstc\n  %dstn = getelementptr i8, ptr %dstc, i64 1\n  call void @llvm.memcpy.p0.p0.i64(ptr %dstn, ptr %pname2, i64 %nl64b, i1 false)\n  %dstn2 = getelementptr i8, ptr %dstn, i64 %nl64b\n  store i8 34, ptr %dstn2\n  %dstcol = getelementptr i8, ptr %dstn2, i64 1\n  store i8 58, ptr %dstcol\n  %dstv = getelementptr i8, ptr %dstcol, i64 1\n  %vpp2 = getelementptr ptr, ptr %vals, i64 %j\n  %vhp2 = load ptr, ptr %vpp2, align 8\n  %isk2_i = icmp eq i8 %kind2, 105\n  %isk2_d = icmp eq i8 %kind2, 100\n  %isk2_b = icmp eq i8 %kind2, 98\n  %isk2_s = icmp eq i8 %kind2, 115\n  br i1 %isk2_i, label %emit_i, label %echk_d\n"
    s += "emit_i:\n  %ival2 = load i64, ptr %vhp2, align 8\n  %bufi = call ptr @ami_rt_i64_to_json(i64 %ival2, i32* %tmpout)\n  %len32i = load i32, ptr %tmpout, align 4\n  %len64i = zext i32 %len32i to i64\n  call void @llvm.memcpy.p0.p0.i64(ptr %dstv, ptr %bufi, i64 %len64i, i1 false)\n  call void @free(ptr %bufi)\n  %nextdst_i = getelementptr i8, ptr %dstv, i64 %len64i\n  br label %emitadv\n"
    s += "echk_d:\n  br i1 %isk2_d, label %emit_d, label %echk_b\n"
    s += "emit_d:\n  %dval2 = load double, ptr %vhp2, align 8\n  %bufd = call ptr @ami_rt_double_to_json(double %dval2, i32* %tmpout)\n  %len32d = load i32, ptr %tmpout, align 4\n  %len64d = zext i32 %len32d to i64\n  call void @llvm.memcpy.p0.p0.i64(ptr %dstv, ptr %bufd, i64 %len64d, i1 false)\n  call void @free(ptr %bufd)\n  %nextdst_d = getelementptr i8, ptr %dstv, i64 %len64d\n  br label %emitadv\n"
    s += "echk_b:\n  br i1 %isk2_b, label %emit_b, label %echk_s\n"
    s += "emit_b:\n  %b82 = load i8, ptr %vhp2, align 1\n  %b12 = icmp ne i8 %b82, 0\n  %bufb = call ptr @ami_rt_bool_to_json(i1 %b12, i32* %tmpout)\n  %len32b = load i32, ptr %tmpout, align 4\n  %len64b = zext i32 %len32b to i64\n  call void @llvm.memcpy.p0.p0.i64(ptr %dstv, ptr %bufb, i64 %len64b, i1 false)\n  call void @free(ptr %bufb)\n  %nextdst_b = getelementptr i8, ptr %dstv, i64 %len64b\n  br label %emitadv\n"
    s += "echk_s:\n  br i1 %isk2_s, label %emit_s, label %emit_o\n"
    s += "emit_s:\n  %bufs = call ptr @ami_rt_string_to_json(ptr %vhp2, i32* %tmpout)\n  %len32s = load i32, ptr %tmpout, align 4\n  %len64s = zext i32 %len32s to i64\n  call void @llvm.memcpy.p0.p0.i64(ptr %dstv, ptr %bufs, i64 %len64s, i1 false)\n  call void @free(ptr %bufs)\n  %nextdst_s = getelementptr i8, ptr %dstv, i64 %len64s\n  br label %emitadv\n"
    s += "emit_o:\n  %vlen2 = call i64 @ami_rt_owned_len(ptr %vhp2)\n  %vptr2 = call ptr @ami_rt_owned_ptr(ptr %vhp2)\n  call void @llvm.memcpy.p0.p0.i64(ptr %dstv, ptr %vptr2, i64 %vlen2, i1 false)\n  %nextdst_o = getelementptr i8, ptr %dstv, i64 %vlen2\n  br label %emitadv\n"
    s += "emitadv:\n  %jnext = add i64 %j, 1\n  br label %emitcont\n"
    s += "emitcont:\n  %nextdst = phi ptr [ %nextdst_i, %emit_i ], [ %nextdst_d, %emit_d ], [ %nextdst_b, %emit_b ], [ %nextdst_s, %emit_s ], [ %nextdst_o, %emit_o ]\n  br label %emitloop\n"
    s += "after:\n  store i8 125, ptr %dst\n  %dptr = ptrtoint ptr %dst to i64\n  %bptr = ptrtoint ptr %buf to i64\n  %delta = sub i64 %dptr, %bptr\n  %olen64 = add i64 %delta, 1\n  %olen = trunc i64 %olen64 to i32\n  store i32 %olen, ptr %outlen, align 4\n  ret ptr %buf\n}\n\n"
    // Array walker: descriptor 'A' + i32 count; if count==0, value handle points to {i64 count; ptr elems}; descriptor carries elem kind at +5
    s += "define ptr @ami_rt_array_to_json(ptr %td, i64 %vh, i32* %outlen) {\n"
    s += "entry:\n  %cntp.i8 = getelementptr i8, ptr %td, i64 1\n  %cntp = bitcast ptr %cntp.i8 to ptr\n  %cnt = load i32, ptr %cntp, align 1\n  %kindp = getelementptr i8, ptr %td, i64 5\n  %ekind = load i8, ptr %kindp, align 1\n  %isZero = icmp eq i32 %cnt, 0\n  br i1 %isZero, label %hdrpath, label %fixed\n"
    s += "hdrpath:\n  %hdr = inttoptr i64 %vh to ptr\n  %n64_hdr = load i64, ptr %hdr, align 8\n  %arrptr.i8 = getelementptr i8, ptr %hdr, i64 8\n  %arrptr = bitcast ptr %arrptr.i8 to ptr\n  %vals_hdr = load ptr, ptr %arrptr, align 8\n  br label %calc\n"
    s += "fixed:\n  %n64_fixed = zext i32 %cnt to i64\n  %vals_fixed = inttoptr i64 %vh to ptr\n  br label %calc\n"
    s += "calc:\n  %n64 = phi i64 [ %n64_hdr, %hdrpath ], [ %n64_fixed, %fixed ]\n  %vals = phi ptr [ %vals_hdr, %hdrpath ], [ %vals_fixed, %fixed ]\n  %has = icmp ugt i64 %n64, 0\n  %nm1 = add i64 %n64, 18446744073709551615\n  %comma64 = select i1 %has, i64 %nm1, i64 0\n  %sum0 = add i64 2, %comma64\n  %tmpout = alloca i32, align 4\n  br label %llen\n"
    s += "llen:\n  %i = phi i64 [ 0, %calc ], [ %inext, %llen ]\n  %sum = phi i64 [ %sum0, %calc ], [ %sum2, %llen ]\n  %done = icmp uge i64 %i, %n64\n  br i1 %done, label %alloc, label %lb\n"
    s += "lb:\n  %vpp = getelementptr ptr, ptr %vals, i64 %i\n  %vhp = load ptr, ptr %vpp, align 8\n  ; compute element len by kind\n  %isk_i = icmp eq i8 %ekind, 105\n  %isk_d = icmp eq i8 %ekind, 100\n  %isk_b = icmp eq i8 %ekind, 98\n  %isk_s = icmp eq i8 %ekind, 115\n  br i1 %isk_i, label %len_i, label %chk_d\n"
    s += "len_i:\n  %ival = load i64, ptr %vhp, align 8\n  %tbi = call ptr @ami_rt_i64_to_json(i64 %ival, i32* %tmpout)\n  %vlen32i = load i32, ptr %tmpout, align 4\n  %vleni = zext i32 %vlen32i to i64\n  call void @free(ptr %tbi)\n  br label %len_accum\n"
    s += "chk_d:\n  br i1 %isk_d, label %len_d, label %chk_b\n"
    s += "len_d:\n  %dval = load double, ptr %vhp, align 8\n  %tbd = call ptr @ami_rt_double_to_json(double %dval, i32* %tmpout)\n  %vlen32d = load i32, ptr %tmpout, align 4\n  %vlend = zext i32 %vlen32d to i64\n  call void @free(ptr %tbd)\n  br label %len_accum\n"
    s += "chk_b:\n  br i1 %isk_b, label %len_b, label %chk_s\n"
    s += "len_b:\n  %b8 = load i8, ptr %vhp, align 1\n  %b1 = icmp ne i8 %b8, 0\n  %tbb = call ptr @ami_rt_bool_to_json(i1 %b1, i32* %tmpout)\n  %vlen32b = load i32, ptr %tmpout, align 4\n  %vlenb = zext i32 %vlen32b to i64\n  call void @free(ptr %tbb)\n  br label %len_accum\n"
    s += "chk_s:\n  br i1 %isk_s, label %len_s, label %len_o\n"
    s += "len_s:\n  %tbs = call ptr @ami_rt_string_to_json(ptr %vhp, i32* %tmpout)\n  %vlen32s = load i32, ptr %tmpout, align 4\n  %vlens = zext i32 %vlen32s to i64\n  call void @free(ptr %tbs)\n  br label %len_accum\n"
    s += "len_o:\n  %vleno = call i64 @ami_rt_owned_len(ptr %vhp)\n  br label %len_accum\n"
    s += "len_accum:\n  %vlen = phi i64 [ %vleni, %len_i ], [ %vlend, %len_d ], [ %vlenb, %len_b ], [ %vlens, %len_s ], [ %vleno, %len_o ]\n  br label %len_accum2\n"
    s += "len_accum2:\n  %sum2 = add i64 %sum, %vlen\n  %inext = add i64 %i, 1\n  br label %llen\n"
    s += "alloc:\n  %buf = call ptr @malloc(i64 %sum)\n  store i8 91, ptr %buf\n  %dst = getelementptr i8, ptr %buf, i64 1\n  br label %emit\n"
    s += "emit:\n  %j = phi i64 [ 0, %alloc ], [ %jnext, %emit2 ]\n  %dstp = phi ptr [ %dst, %alloc ], [ %nextdst, %emit2 ]\n  %done2 = icmp uge i64 %j, %n64\n  br i1 %done2, label %end, label %emit2\n"
    s += "emit2:\n  %gt0 = icmp ugt i64 %j, 0\n  br i1 %gt0, label %comma, label %noc\n"
    s += "comma:\n  store i8 44, ptr %dstp\n  %dstc_c2 = getelementptr i8, ptr %dstp, i64 1\n  br label %nocont\n"
    s += "noc:\n  %dstc_n2 = getelementptr i8, ptr %dstp, i64 0\n  br label %nocont\n"
    s += "nocont:\n  %dstc = phi ptr [ %dstc_c2, %comma ], [ %dstc_n2, %noc ]\n  %vpp2 = getelementptr ptr, ptr %vals, i64 %j\n  %vhp2 = load ptr, ptr %vpp2, align 8\n  ; emit by kind\n  %isk2_i = icmp eq i8 %ekind, 105\n  %isk2_d = icmp eq i8 %ekind, 100\n  %isk2_b = icmp eq i8 %ekind, 98\n  %isk2_s = icmp eq i8 %ekind, 115\n  br i1 %isk2_i, label %emit_i, label %echk_d\n"
    s += "emit_i:\n  %ival2 = load i64, ptr %vhp2, align 8\n  %bufi = call ptr @ami_rt_i64_to_json(i64 %ival2, i32* %tmpout)\n  %len32i = load i32, ptr %tmpout, align 4\n  %len64i = zext i32 %len32i to i64\n  call void @llvm.memcpy.p0.p0.i64(ptr %dstc, ptr %bufi, i64 %len64i, i1 false)\n  call void @free(ptr %bufi)\n  %nextdst_i = getelementptr i8, ptr %dstc, i64 %len64i\n  br label %emitadv\n"
    s += "echk_d:\n  br i1 %isk2_d, label %emit_d, label %echk_b\n"
    s += "emit_d:\n  %dval2 = load double, ptr %vhp2, align 8\n  %bufd = call ptr @ami_rt_double_to_json(double %dval2, i32* %tmpout)\n  %len32d = load i32, ptr %tmpout, align 4\n  %len64d = zext i32 %len32d to i64\n  call void @llvm.memcpy.p0.p0.i64(ptr %dstc, ptr %bufd, i64 %len64d, i1 false)\n  call void @free(ptr %bufd)\n  %nextdst_d = getelementptr i8, ptr %dstc, i64 %len64d\n  br label %emitadv\n"
    s += "echk_b:\n  br i1 %isk2_b, label %emit_b, label %echk_s\n"
    s += "emit_b:\n  %b82 = load i8, ptr %vhp2, align 1\n  %b12 = icmp ne i8 %b82, 0\n  %bufb = call ptr @ami_rt_bool_to_json(i1 %b12, i32* %tmpout)\n  %len32b = load i32, ptr %tmpout, align 4\n  %len64b = zext i32 %len32b to i64\n  call void @llvm.memcpy.p0.p0.i64(ptr %dstc, ptr %bufb, i64 %len64b, i1 false)\n  call void @free(ptr %bufb)\n  %nextdst_b = getelementptr i8, ptr %dstc, i64 %len64b\n  br label %emitadv\n"
    s += "echk_s:\n  br i1 %isk2_s, label %emit_s, label %emit_o\n"
    s += "emit_s:\n  %bufs = call ptr @ami_rt_string_to_json(ptr %vhp2, i32* %tmpout)\n  %len32s = load i32, ptr %tmpout, align 4\n  %len64s = zext i32 %len32s to i64\n  call void @llvm.memcpy.p0.p0.i64(ptr %dstc, ptr %bufs, i64 %len64s, i1 false)\n  call void @free(ptr %bufs)\n  %nextdst_s = getelementptr i8, ptr %dstc, i64 %len64s\n  br label %emitadv\n"
    s += "emit_o:\n  %vlen2 = call i64 @ami_rt_owned_len(ptr %vhp2)\n  %vptr2 = call ptr @ami_rt_owned_ptr(ptr %vhp2)\n  call void @llvm.memcpy.p0.p0.i64(ptr %dstc, ptr %vptr2, i64 %vlen2, i1 false)\n  %nextdst_o = getelementptr i8, ptr %dstc, i64 %vlen2\n  br label %emitadv\n"
    s += "emitadv:\n  %jnext = add i64 %j, 1\n  br label %emitcont\n"
    s += "emitcont:\n  %nextdst = phi ptr [ %nextdst_i, %emit_i ], [ %nextdst_d, %emit_d ], [ %nextdst_b, %emit_b ], [ %nextdst_s, %emit_s ], [ %nextdst_o, %emit_o ]\n  br label %emit\n"
    s += "end:\n  store i8 93, ptr %dstp\n  %dptr = ptrtoint ptr %dstp to i64\n  %bptr = ptrtoint ptr %buf to i64\n  %delta = sub i64 %dptr, %bptr\n  %olen64 = add i64 %delta, 1\n  %olen = trunc i64 %olen64 to i32\n  store i32 %olen, ptr %outlen, align 4\n  ret ptr %buf\n}\n\n"
    // Central dispatcher
    s += "define ptr @ami_rt_value_to_json(ptr %typedescr, i64 %vh, i32* %outlen) {\n"
    s += "entry:\n  %k = load i8, ptr %typedescr, align 1\n  %isS = icmp eq i8 %k, 83\n  br i1 %isS, label %struct, label %chkA\n"
    s += "struct:\n  %js1 = call ptr @ami_rt_struct_to_json(ptr %typedescr, i64 %vh, i32* %outlen)\n  ret ptr %js1\n"
    s += "chkA:\n  %isA = icmp eq i8 %k, 65\n  br i1 %isA, label %arr, label %fallback\n"
    s += "arr:\n  %js2 = call ptr @ami_rt_array_to_json(ptr %typedescr, i64 %vh, i32* %outlen)\n  ret ptr %js2\n"
    s += "fallback:\n  %p = inttoptr i64 %vh to ptr\n  %js3 = call ptr @ami_rt_structured_to_json(ptr %p, i32* %outlen)\n  ret ptr %js3\n}\n\n"
    // Helper: parse Event payload as signed int64 (ASCII decimal), ignoring leading spaces.
    s += "define i64 @ami_rt_event_payload_to_i64(ptr %ev) {\n"
    s += "entry:\n  %pfield = bitcast ptr %ev to ptr\n  %pp = load ptr, ptr %pfield, align 8\n  %lenptr.i8 = getelementptr i8, ptr %ev, i64 8\n  %lfield = bitcast ptr %lenptr.i8 to ptr\n  %plen = load i64, ptr %lfield, align 8\n  br label %scan\n"
    s += "scan:\n  %i = phi i64 [ 0, %entry ], [ %inext, %scan ]\n  %done = icmp uge i64 %i, %plen\n  br i1 %done, label %ret0, label %scan2\n"
    s += "scan2:\n  %addr = getelementptr i8, ptr %pp, i64 %i\n  %b = load i8, ptr %addr, align 1\n  %isSp = icmp eq i8 %b, 32\n  %isNl = icmp eq i8 %b, 10\n  %isCr = icmp eq i8 %b, 13\n  %isTb = icmp eq i8 %b, 9\n  %sp1 = or i1 %isSp, %isNl\n  %sp2 = or i1 %isCr, %isTb\n  %isWS = or i1 %sp1, %sp2\n  br i1 %isWS, label %scanadv, label %signcheck\n"
    s += "scanadv:\n  %inext = add i64 %i, 1\n  br label %scan\n"
    s += "signcheck:\n  %addr_s = getelementptr i8, ptr %pp, i64 %i\n  %bs = load i8, ptr %addr_s, align 1\n  %isMinus = icmp eq i8 %bs, 45\n  br i1 %isMinus, label %after_sign, label %parse\n"
    s += "after_sign:\n  %i2 = add i64 %i, 1\n  br label %parse_start\n"
    s += "parse:\n  br label %parse_start\n"
    s += "parse_start:\n  %j = phi i64 [ %i, %parse ], [ %i2, %after_sign ]\n  %neg = phi i1 [ false, %parse ], [ true, %after_sign ]\n  %acc = phi i64 [ 0, %parse ], [ 0, %after_sign ]\n  br label %loop\n"
    s += "loop:\n  %end = icmp uge i64 %j, %plen\n  br i1 %end, label %finish, label %body\n"
    s += "body:\n  %addr2 = getelementptr i8, ptr %pp, i64 %j\n  %b2 = load i8, ptr %addr2, align 1\n  %ge0 = icmp sge i8 %b2, 48\n  %le9 = icmp sle i8 %b2, 57\n  %isd = and i1 %ge0, %le9\n  br i1 %isd, label %accum, label %finish\n"
    s += "accum:\n  %v = zext i8 %b2 to i64\n  %d = sub i64 %v, 48\n  %acc10 = mul i64 %acc, 10\n  %acc2 = add i64 %acc10, %d\n  %j2 = add i64 %j, 1\n  br label %loop\n"
    s += "finish:\n  %accv = phi i64 [ %acc, %loop ], [ %acc2, %accum ]\n  %negv = phi i1 [ %neg, %loop ], [ %neg, %accum ]\n  %negval = sub i64 0, %accv\n  %sel = select i1 %negv, i64 %negval, i64 %accv\n  ret i64 %sel\n"
    s += "ret0:\n  ret i64 0\n}\n\n"

    // Helper: parse Event payload as double (ASCII decimal with optional fraction)
    s += "define double @ami_rt_event_payload_to_double(ptr %ev) {\n"
    s += "entry:\n  %pfield = bitcast ptr %ev to ptr\n  %pp = load ptr, ptr %pfield, align 8\n  %lenptr.i8 = getelementptr i8, ptr %ev, i64 8\n  %lfield = bitcast ptr %lenptr.i8 to ptr\n  %plen = load i64, ptr %lfield, align 8\n  br label %scan\n"
    s += "scan:\n  %i = phi i64 [ 0, %entry ], [ %inext, %scan ]\n  %done = icmp uge i64 %i, %plen\n  br i1 %done, label %ret0d, label %scan2\n"
    s += "scan2:\n  %addr = getelementptr i8, ptr %pp, i64 %i\n  %b = load i8, ptr %addr, align 1\n  %isSp = icmp eq i8 %b, 32\n  %isTb = icmp eq i8 %b, 9\n  %isWS = or i1 %isSp, %isTb\n  br i1 %isWS, label %scanadv, label %signcheck\n"
    s += "scanadv:\n  %inext = add i64 %i, 1\n  br label %scan\n"
    s += "signcheck:\n  %addr_s = getelementptr i8, ptr %pp, i64 %i\n  %bs = load i8, ptr %addr_s, align 1\n  %isMinus = icmp eq i8 %bs, 45\n  br i1 %isMinus, label %after_sign, label %parse\n"
    s += "after_sign:\n  %i2 = add i64 %i, 1\n  br label %parse_start\n"
    s += "parse:\n  br label %parse_start\n"
    s += "parse_start:\n  %j = phi i64 [ %i, %parse ], [ %i2, %after_sign ]\n  %neg = phi i1 [ false, %parse ], [ true, %after_sign ]\n  %acc = phi double [ 0.0, %parse ], [ 0.0, %after_sign ]\n  br label %intloop\n"
    s += "intloop:\n  %end = icmp uge i64 %j, %plen\n  br i1 %end, label %fracchk, label %ibody\n"
    s += "ibody:\n  %addr2 = getelementptr i8, ptr %pp, i64 %j\n  %b2 = load i8, ptr %addr2, align 1\n  %ge0 = icmp sge i8 %b2, 48\n  %le9 = icmp sle i8 %b2, 57\n  %isd = and i1 %ge0, %le9\n  br i1 %isd, label %iaccum, label %fracchk\n"
    s += "iaccum:\n  %v = zext i8 %b2 to i64\n  %vdiff = sub i64 %v, 48\n  %d = sitofp i64 %vdiff to double\n  %acc10 = fmul double %acc, 10.0\n  %acc2 = fadd double %acc10, %d\n  %j2 = add i64 %j, 1\n  br label %intloop\n"
    s += "fracchk:\n  %addr3 = getelementptr i8, ptr %pp, i64 %j\n  %b3 = load i8, ptr %addr3, align 1\n  %isDot = icmp eq i8 %b3, 46\n  br i1 %isDot, label %fracstart, label %finishd\n"
    s += "fracstart:\n  %k = add i64 %j, 1\n  %frac = phi double [ 0.0, %fracstart ], [ 0.0, %fracstart ]\n  %scale = phi double [ 1.0, %fracstart ], [ 1.0, %fracstart ]\n  br label %floop\n"
    s += "floop:\n  %fend = icmp uge i64 %k, %plen\n  br i1 %fend, label %finishd, label %fbody\n"
    s += "fbody:\n  %addr4 = getelementptr i8, ptr %pp, i64 %k\n  %b4 = load i8, ptr %addr4, align 1\n  %ge0f = icmp sge i8 %b4, 48\n  %le9f = icmp sle i8 %b4, 57\n  %isdf = and i1 %ge0f, %le9f\n  br i1 %isdf, label %faccum, label %finishd\n"
    s += "faccum:\n  %vf = zext i8 %b4 to i64\n  %vfdiff = sub i64 %vf, 48\n  %df = sitofp i64 %vfdiff to double\n  %scale2 = fmul double %scale, 10.0\n  %divv = fdiv double %df, %scale2\n  %frac2 = fadd double %frac, %divv\n  %k2 = add i64 %k, 1\n  br label %floop\n"
    s += "finishd:\n  %accv = phi double [ %acc, %fracchk ], [ %acc2, %iaccum ]\n  %negv = phi i1 [ %neg, %fracchk ], [ %neg, %iaccum ]\n  %fracv = phi double [ 0.0, %fracchk ], [ %frac2, %faccum ]\n  %sum = fadd double %accv, %fracv\n  %one = fadd double 0.0, 1.0\n  %neg1 = fsub double 0.0, 1.0\n  %sign = select i1 %negv, double %neg1, double %one\n  %res = fmul double %sum, %sign\n  ret double %res\n"
    s += "ret0d:\n  ret double 0.0\n}\n\n"

    // Helper: parse Event payload as boolean
    s += "define i1 @ami_rt_event_payload_to_bool(ptr %ev) {\n"
    s += "entry:\n  %pfield = bitcast ptr %ev to ptr\n  %pp = load ptr, ptr %pfield, align 8\n  %b0 = load i8, ptr %pp, align 1\n  %isT = icmp eq i8 %b0, 116\n  %val = select i1 %isT, i1 true, i1 false\n  ret i1 %val\n}\n\n"

    // Helper: return pointer to payload as string pointer (not NUL-terminated guaranteed)
    s += "define ptr @ami_rt_event_payload_to_string(ptr %ev) {\n"
    s += "entry:\n  %pfield = bitcast ptr %ev to ptr\n  %pp = load ptr, ptr %pfield, align 8\n  ret ptr %pp\n}\n\n"

    // Structured JSON field readers using libc helpers for searching and parsing.
    s += "declare ptr @strstr(ptr, ptr)\n"
    s += "declare i64 @strtoll(ptr, ptr, i32)\n"
    s += "declare double @strtod(ptr, ptr)\n\n"

    // Duplicate payload into NUL-terminated buffer
    s += "define ptr @ami_rt_dup_nul(ptr %p, i64 %n) {\n"
    s += "entry:\n  %sz = add i64 %n, 1\n  %buf = call ptr @malloc(i64 %sz)\n  call void @llvm.memcpy.p0.p0.i64(ptr %buf, ptr %p, i64 %n, i1 false)\n  %last = getelementptr i8, ptr %buf, i64 %n\n  store i8 0, ptr %last, align 1\n  ret ptr %buf\n}\n\n"

    // Build quoted key string "<seg>" as C-string
    s += "define ptr @ami_rt_build_quoted(ptr %seg, i32 %len) {\n"
    s += "entry:\n  %n = zext i32 %len to i64\n  %sz = add i64 %n, 3\n  %buf = call ptr @malloc(i64 %sz)\n  store i8 34, ptr %buf, align 1\n  %dst = getelementptr i8, ptr %buf, i64 1\n  call void @llvm.memcpy.p0.p0.i64(ptr %dst, ptr %seg, i64 %n, i1 false)\n  %n1 = add i64 %n, 1\n  %q2 = getelementptr i8, ptr %buf, i64 %n1\n  store i8 34, ptr %q2, align 1\n  %n2 = add i64 %n, 2\n  %nul = getelementptr i8, ptr %buf, i64 %n2\n  store i8 0, ptr %nul, align 1\n  ret ptr %buf\n}\n\n"

    // Whitespace skipper used by multiple scanners
    s += "define ptr @ami_rt_skip_ws(ptr %p) {\n"
    s += "entry:\n  br label %ws\n"
    s += "ws:\n  %cur = phi ptr [ %p, %entry ], [ %p2, %ws_adv ]\n  %c = load i8, ptr %cur, align 1\n  %isSp = icmp eq i8 %c, 32\n  %isTb = icmp eq i8 %c, 9\n  %isNl = icmp eq i8 %c, 10\n  %isCr = icmp eq i8 %c, 13\n  %sp1 = or i1 %isSp, %isTb\n  %sp2 = or i1 %isNl, %isCr\n  %isWS = or i1 %sp1, %sp2\n  br i1 %isWS, label %ws_adv, label %done\n"
    s += "ws_adv:\n  %p2 = getelementptr i8, ptr %cur, i64 1\n  br label %ws\n"
    s += "done:\n  ret ptr %cur\n}\n\n"

    // Advance to first non-space after ':' starting at p
    s += "define ptr @ami_rt_after_colon(ptr %p) {\n"
    s += "entry:\n  br label %loop\n"
    s += "loop:\n  %cur = phi ptr [ %p, %entry ], [ %p2, %next ]\n  %c = load i8, ptr %cur, align 1\n  %isColon = icmp eq i8 %c, 58\n  br i1 %isColon, label %adv, label %next\n"
    s += "next:\n  %p2 = getelementptr i8, ptr %cur, i64 1\n  br label %loop\n"
    s += "adv:\n  %p3 = getelementptr i8, ptr %cur, i64 1\n  %p4 = call ptr @ami_rt_skip_ws(ptr %p3)\n  ret ptr %p4\n}\n\n"

    // Scan to the end of a string starting at an opening quote at %p; returns ptr after closing quote
    s += "define ptr @ami_rt_scan_string_end(ptr %p) {\n"
    s += "entry:\n  %start = getelementptr i8, ptr %p, i64 1\n  br label %loop\n"
    s += "loop:\n  %cur = phi ptr [ %start, %entry ], [ %next2, %esc ], [ %next1, %adv ]\n  %c = load i8, ptr %cur, align 1\n  %isEsc = icmp eq i8 %c, 92\n  br i1 %isEsc, label %esc, label %chkq\n"
    s += "esc:\n  %next2 = getelementptr i8, ptr %cur, i64 2\n  br label %loop\n"
    s += "chkq:\n  %isQ = icmp eq i8 %c, 34\n  br i1 %isQ, label %done, label %adv\n"
    s += "adv:\n  %next1 = getelementptr i8, ptr %cur, i64 1\n  br label %loop\n"
    s += "done:\n  %aft = getelementptr i8, ptr %cur, i64 1\n  ret ptr %aft\n}\n\n"

    // (Removed robust object key scanner; object-case uses quoted-key strstr + after_colon)

    // Return pointer to the idx-th element within a JSON array at %arr
    s += "define ptr @ami_rt_array_index(ptr %arr, i32 %idx) {\n"
    s += "entry:\n  %p0 = call ptr @ami_rt_skip_ws(ptr %arr)\n  %c0 = load i8, ptr %p0, align 1\n  %isLB = icmp eq i8 %c0, 91\n  br i1 %isLB, label %start, label %nf\n"
    s += "start:\n  %q0 = getelementptr i8, ptr %p0, i64 1\n  br label %loop\n"
    s += "loop:\n  %pos = phi ptr [ %q0, %start ], [ %nextpos, %bump ]\n  %idxcur = phi i32 [ 0, %start ], [ %j1, %bump ]\n  %ws = call ptr @ami_rt_skip_ws(ptr %pos)\n  %c = load i8, ptr %ws, align 1\n  %isEnd = icmp eq i8 %c, 93\n  br i1 %isEnd, label %nf, label %elem\n"
    s += "elem:\n  %match = icmp eq i32 %idxcur, %idx\n  br i1 %match, label %retp, label %skipelem\n"
    s += "retp:\n  ret ptr %ws\n"
    s += "skipelem:\n  ; skip current element until comma at depth 0 (respect nested structures and strings)\n  br label %sloop\n"
    s += "sloop:\n  %h = phi ptr [ %ws, %skipelem ], [ %hn, %scont ]\n  %depth = phi i32 [ 0, %skipelem ], [ %d4, %scont ]\n  %ch = load i8, ptr %h, align 1\n  %isq = icmp eq i8 %ch, 34\n  br i1 %isq, label %sstr, label %sbody\n"
    s += "sstr:\n  %h1 = call ptr @ami_rt_scan_string_end(ptr %h)\n  br label %scont\n"
    s += "sbody:\n  %iso = icmp eq i8 %ch, 123\n  %isc = icmp eq i8 %ch, 125\n  %isa = icmp eq i8 %ch, 91\n  %isz = icmp eq i8 %ch, 93\n  %inc_o = select i1 %iso, i32 1, i32 0\n  %inc_a = select i1 %isa, i32 1, i32 0\n  %dec_o = select i1 %isc, i32 1, i32 0\n  %dec_a = select i1 %isz, i32 1, i32 0\n  %d1 = add i32 %depth, %inc_o\n  %d2 = add i32 %d1, %inc_a\n  %d3 = sub i32 %d2, %dec_o\n  %d4 = sub i32 %d3, %dec_a\n  %hinc = getelementptr i8, ptr %h, i64 1\n  br label %scont\n"
    s += "scont:\n  %hn = phi ptr [ %h1, %sstr ], [ %hinc, %sbody ]\n  %dcur = phi i32 [ %depth, %sstr ], [ %d4, %sbody ]\n  %dz = icmp eq i32 %dcur, 0\n  %isComma = icmp eq i8 %ch, 44\n  %atComma = and i1 %dz, %isComma\n  br i1 %atComma, label %bump, label %sloop\n"
    s += "bump:\n  %j1 = add i32 %idxcur, 1\n  %nextpos = getelementptr i8, ptr %hn, i64 1\n  br label %loop\n"
    s += "nf:\n  ret ptr null\n}\n\n"

    // Find pointer to value for dotted path within JSON buffer; supports array indices and escape-aware scanning
    s += "define ptr @ami_rt_find_path(ptr %json, ptr %path, i32 %plen) {\n"
    s += "entry:\n  br label %segloop\n"
    s += "segloop:\n  %p = phi ptr [ %path, %entry ], [ %pnext, %contnest ]\n  %r = phi i32 [ %plen, %entry ], [ %rnext, %contnest ]\n  %b = phi ptr [ %json, %entry ], [ %bnext, %contnest ]\n  %i = alloca i32, align 4\n  %isnum = alloca i1, align 1\n  store i32 0, ptr %i, align 4\n  store i1 true, ptr %isnum, align 1\n  br label %finddot\n"
    s += "finddot:\n  %idx = load i32, ptr %i, align 4\n  %done = icmp uge i32 %idx, %r\n  br i1 %done, label %segfound, label %fdstep\n"
    s += "fdstep:\n  %cp = getelementptr i8, ptr %p, i32 %idx\n  %ch = load i8, ptr %cp, align 1\n  %isdot = icmp eq i8 %ch, 46\n  %ge0 = icmp sge i8 %ch, 48\n  %le9 = icmp sle i8 %ch, 57\n  %isdig = and i1 %ge0, %le9\n  %numcur = load i1, ptr %isnum, align 1\n  %numupd = and i1 %numcur, %isdig\n  store i1 %numupd, ptr %isnum, align 1\n  br i1 %isdot, label %segfound, label %fdcont\n"
    s += "fdcont:\n  %idx2 = add i32 %idx, 1\n  store i32 %idx2, ptr %i, align 4\n  br label %finddot\n"
    s += "segfound:\n  %seglen = load i32, ptr %i, align 4\n  %isN = load i1, ptr %isnum, align 1\n  br i1 %isN, label %arrcase, label %objcase\n"
    s += "arrcase:\n  ; parse numeric index from path segment\n  %acc = alloca i32, align 4\n  store i32 0, ptr %acc, align 4\n  %k = alloca i32, align 4\n  store i32 0, ptr %k, align 4\n  br label %iparse\n"
    s += "iparse:\n  %kcur = load i32, ptr %k, align 4\n  %kend = icmp uge i32 %kcur, %seglen\n  br i1 %kend, label %iget, label %inext\n"
    s += "inext:\n  %pp = getelementptr i8, ptr %p, i32 %kcur\n  %ch2 = load i8, ptr %pp, align 1\n  %ch2z = zext i8 %ch2 to i32\n  %d = sub i32 %ch2z, 48\n  %old = load i32, ptr %acc, align 4\n  %tmp = mul i32 %old, 10\n  %new = add i32 %tmp, %d\n  store i32 %new, ptr %acc, align 4\n  %k2 = add i32 %kcur, 1\n  store i32 %k2, ptr %k, align 4\n  br label %iparse\n"
    s += "iget:\n  %index = load i32, ptr %acc, align 4\n  %val_arr = call ptr @ami_rt_array_index(ptr %b, i32 %index)\n  %atend = icmp eq i32 %seglen, %r\n  br i1 %atend, label %ret_arr, label %contnest\n"
    s += "ret_arr:\n  ret ptr %val_arr\n"
    s += "objcase:\n  %qkey = call ptr @ami_rt_build_quoted(ptr %p, i32 %seglen)\n  %hit = call ptr @strstr(ptr %b, ptr %qkey)\n  %isnull2 = icmp eq ptr %hit, null\n  br i1 %isnull2, label %notfound, label %afterkey2\n"
    s += "afterkey2:\n  %seg64 = zext i32 %seglen to i64\n  %two = add i64 0, 2\n  %off = add i64 %seg64, %two\n  %afterq = getelementptr i8, ptr %hit, i64 %off\n  %val_obj = call ptr @ami_rt_after_colon(ptr %afterq)\n  %atend2 = icmp eq i32 %seglen, %r\n  br i1 %atend2, label %ret_obj, label %contnest\n"
    s += "ret_obj:\n  ret ptr %val_obj\n"
    s += "contnest:\n  %bnext = phi ptr [ %val_arr, %iget ], [ %val_obj, %afterkey2 ]\n  %segp = add i32 %seglen, 1\n  %pnext = getelementptr i8, ptr %p, i32 %segp\n  %rnext = sub i32 %r, %segp\n  br label %segloop\n"
    s += "notfound:\n  ret ptr null\n}\n\n"

    s += "define i64 @ami_rt_event_get_i64(ptr %ev, ptr %path, i32 %plen) {\n"
    s += "entry:\n  %pfield = bitcast ptr %ev to ptr\n  %pp = load ptr, ptr %pfield, align 8\n  %lenptr.i8 = getelementptr i8, ptr %ev, i64 8\n  %lfield = bitcast ptr %lenptr.i8 to ptr\n  %plen64 = load i64, ptr %lfield, align 8\n  %cpy = call ptr @ami_rt_dup_nul(ptr %pp, i64 %plen64)\n  %v = call ptr @ami_rt_find_path(ptr %cpy, ptr %path, i32 %plen)\n  %isnull = icmp eq ptr %v, null\n  br i1 %isnull, label %ret0, label %parse\n"
    s += "parse:\n  %n = call i64 @strtoll(ptr %v, ptr null, i32 10)\n  call void @free(ptr %cpy)\n  ret i64 %n\n"
    s += "ret0:\n  call void @free(ptr %cpy)\n  ret i64 0\n}\n\n"

    s += "define double @ami_rt_event_get_double(ptr %ev, ptr %path, i32 %plen) {\n"
    s += "entry:\n  %pfield = bitcast ptr %ev to ptr\n  %pp = load ptr, ptr %pfield, align 8\n  %lenptr.i8 = getelementptr i8, ptr %ev, i64 8\n  %lfield = bitcast ptr %lenptr.i8 to ptr\n  %plen64 = load i64, ptr %lfield, align 8\n  %cpy = call ptr @ami_rt_dup_nul(ptr %pp, i64 %plen64)\n  %v = call ptr @ami_rt_find_path(ptr %cpy, ptr %path, i32 %plen)\n  %isnull = icmp eq ptr %v, null\n  br i1 %isnull, label %ret0, label %parse\n"
    s += "parse:\n  %x = call double @strtod(ptr %v, ptr null)\n  call void @free(ptr %cpy)\n  ret double %x\n"
    s += "ret0:\n  call void @free(ptr %cpy)\n  ret double 0.0\n}\n\n"

    s += "define i1 @ami_rt_event_get_bool(ptr %ev, ptr %path, i32 %plen) {\n"
    s += "entry:\n  %pfield = bitcast ptr %ev to ptr\n  %pp = load ptr, ptr %pfield, align 8\n  %lenptr.i8 = getelementptr i8, ptr %ev, i64 8\n  %lfield = bitcast ptr %lenptr.i8 to ptr\n  %plen64 = load i64, ptr %lfield, align 8\n  %cpy = call ptr @ami_rt_dup_nul(ptr %pp, i64 %plen64)\n  %v = call ptr @ami_rt_find_path(ptr %cpy, ptr %path, i32 %plen)\n  %isnull = icmp eq ptr %v, null\n  br i1 %isnull, label %ret0, label %parse\n"
    s += "parse:\n  %c = load i8, ptr %v, align 1\n  %isT = icmp eq i8 %c, 116\n  call void @free(ptr %cpy)\n  ret i1 %isT\n"
    s += "ret0:\n  call void @free(ptr %cpy)\n  ret i1 false\n}\n\n"

    s += "define ptr @ami_rt_event_get_string(ptr %ev, ptr %path, i32 %plen) {\n"
    s += "entry:\n  %pfield = bitcast ptr %ev to ptr\n  %pp = load ptr, ptr %pfield, align 8\n  %lenptr.i8 = getelementptr i8, ptr %ev, i64 8\n  %lfield = bitcast ptr %lenptr.i8 to ptr\n  %plen64 = load i64, ptr %lfield, align 8\n  %cpy = call ptr @ami_rt_dup_nul(ptr %pp, i64 %plen64)\n  %v = call ptr @ami_rt_find_path(ptr %cpy, ptr %path, i32 %plen)\n  ret ptr %v\n}\n\n"
    return s
}

// WriteRuntimeLL writes the runtime LLVM IR text to the given directory and returns the file path.
// file writing moved to runtime_write.go to satisfy single-declaration rule

// gpuAllowedBackends inspects the AMI_GPU_BACKENDS env var and returns booleans for
// allowing metal, cuda, opencl in the generated runtime IR. Empty or missing means all allowed.
// moved to runtime_gpu_backends.go

// itoa is a tiny integer to string formatter for small positive values used in IR lengths.
// itoa helper provided in emit_util.go
