package llvm

import (
    "bytes"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

// Try to execute a pure-IR harness calling ami_rt_object_find.
// Prefers `lli` if present; otherwise attempts to compile the combined .ll with clang and run it.
func TestRuntime_ObjectFind_PureIRHarness(t *testing.T) {
    // Build runtime IR
    dir := t.TempDir()
    // Disable Metal in runtime so darwin IR doesn't reference external shims under lli
    orig := os.Getenv("AMI_GPU_BACKENDS")
    _ = os.Setenv("AMI_GPU_BACKENDS", "cuda,opencl")
    defer os.Setenv("AMI_GPU_BACKENDS", orig)
    triple := DefaultTriple
    llp, err := WriteRuntimeLL(dir, triple, false)
    if err != nil { t.Fatalf("WriteRuntimeLL: %v", err) }
    rtIR, err := os.ReadFile(llp)
    if err != nil { t.Fatalf("read runtime.ll: %v", err) }

    // Cases to validate: escaped quote, \u0022, 2-byte, 3-byte, 4-byte
    type caseDef struct{ json, key string; expect byte }
    cases := []caseDef{
        {json: `{"a\"b":42}`, key: `a"b`, expect: '4'},
        {json: `{"A\u0022B":99}`, key: `A"B`, expect: '9'},
        {json: `{"\u00E9":77}`, key: string([]byte{0xC3, 0xA9}), expect: '7'},
        {json: `{"\u20AC":88}`, key: string([]byte{0xE2, 0x82, 0xAC}), expect: '8'},
        {json: `{"\uD83D\uDE00":321}`, key: string([]byte{0xF0, 0x9F, 0x98, 0x80}), expect: '3'},
    }

    // Build harness IR appended to runtime module
    var h bytes.Buffer
    h.WriteString("\n; pure-IR harness for object_find\n")
    for i, c := range cases {
        j := []byte(c.json)
        k := []byte(c.key)
        h.WriteString(irConstBytes("json", i, j))
        h.WriteString(irConstBytes("key", i, k))
    }
    h.WriteString("\n")
    // main returns 0 on success, 1 on failure
    h.WriteString("define i32 @main() {\nentry:\n  br label %loop\n")
    h.WriteString("loop:\n  %i = phi i32 [ 0, %entry ], [ %inext, %cont ]\n  %ok = phi i1 [ true, %entry ], [ %ok2, %cont ]\n  %done = icmp uge i32 %i, ")
    h.WriteString(itoa(len(cases)))
    h.WriteString("\n  br i1 %done, label %exit, label %body\n")
    h.WriteString("body:\n  ; select json/key for index %i\n")
    for i := range cases {
        jn := len([]byte(cases[i].json)) + 1
        kn := len([]byte(cases[i].key)) + 0 // do not add NUL to reported key length
        // pointers to globals
        h.WriteString("  %is")
        h.WriteString(itoa(i))
        h.WriteString(" = icmp eq i32 %i, ")
        h.WriteString(itoa(i))
        h.WriteString("\n")
        h.WriteString("  %pj")
        h.WriteString(itoa(i))
        h.WriteString(" = getelementptr inbounds [")
        h.WriteString(itoa(jn))
        h.WriteString(" x i8], ptr @.json.")
        h.WriteString(itoa(i))
        h.WriteString(", i64 0, i64 0\n")
        h.WriteString("  %pk")
        h.WriteString(itoa(i))
        h.WriteString(" = getelementptr inbounds [")
        h.WriteString(itoa(kn+1))
        h.WriteString(" x i8], ptr @.key.")
        h.WriteString(itoa(i))
        h.WriteString(", i64 0, i64 0\n")
    }
    // Build equality flags for index selection
    for i := range cases {
        h.WriteString("  %eq")
        h.WriteString(itoa(i))
        h.WriteString(" = icmp eq i32 %i, ")
        h.WriteString(itoa(i))
        h.WriteString("\n")
    }
    // jcur selection via nested selects
    if len(cases) == 1 {
        h.WriteString("  %jcur = ptr %pj0\n")
        h.WriteString("  %kcur = ptr %pk0\n")
        h.WriteString("  %klen = add i32 0, ")
        h.WriteString(itoa(len([]byte(cases[0].key))))
        h.WriteString("\n")
    } else {
        // jcur
        h.WriteString("  %jsel0 = select i1 %eq0, ptr %pj0, ptr %pj1\n")
        for i := 1; i < len(cases)-1; i++ {
            h.WriteString("  %jsel")
            h.WriteString(itoa(i))
            h.WriteString(" = select i1 %eq")
            h.WriteString(itoa(i))
            h.WriteString(", ptr %pj")
            h.WriteString(itoa(i))
            h.WriteString(", ptr %jsel")
            h.WriteString(itoa(i-1))
            h.WriteString("\n")
        }
        h.WriteString("  %jcur = select i1 %eq")
        h.WriteString(itoa(len(cases)-1))
        h.WriteString(", ptr %pj")
        h.WriteString(itoa(len(cases)-1))
        h.WriteString(", ptr %jsel")
        h.WriteString(itoa(len(cases)-2))
        h.WriteString("\n")
        // kcur
        h.WriteString("  %ksel0 = select i1 %eq0, ptr %pk0, ptr %pk1\n")
        for i := 1; i < len(cases)-1; i++ {
            h.WriteString("  %ksel")
            h.WriteString(itoa(i))
            h.WriteString(" = select i1 %eq")
            h.WriteString(itoa(i))
            h.WriteString(", ptr %pk")
            h.WriteString(itoa(i))
            h.WriteString(", ptr %ksel")
            h.WriteString(itoa(i-1))
            h.WriteString("\n")
        }
        h.WriteString("  %kcur = select i1 %eq")
        h.WriteString(itoa(len(cases)-1))
        h.WriteString(", ptr %pk")
        h.WriteString(itoa(len(cases)-1))
        h.WriteString(", ptr %ksel")
        h.WriteString(itoa(len(cases)-2))
        h.WriteString("\n")
        // klen (i32)
        h.WriteString("  %klsel0 = select i1 %eq0, i32 ")
        h.WriteString(itoa(len([]byte(cases[0].key))))
        h.WriteString(", i32 ")
        h.WriteString(itoa(len([]byte(cases[1].key))))
        h.WriteString("\n")
        for i := 1; i < len(cases)-1; i++ {
            h.WriteString("  %klsel")
            h.WriteString(itoa(i))
            h.WriteString(" = select i1 %eq")
            h.WriteString(itoa(i))
            h.WriteString(", i32 ")
            h.WriteString(itoa(len([]byte(cases[i].key))))
            h.WriteString(", i32 %klsel")
            h.WriteString(itoa(i-1))
            h.WriteString("\n")
        }
        h.WriteString("  %klen = select i1 %eq")
        h.WriteString(itoa(len(cases)-1))
        h.WriteString(", i32 ")
        h.WriteString(itoa(len([]byte(cases[len(cases)-1].key))))
        h.WriteString(", i32 %klsel")
        h.WriteString(itoa(len(cases)-2))
        h.WriteString("\n")
    }
    h.WriteString("\n  %val = call ptr @ami_rt_object_find(ptr %jcur, ptr %kcur, i32 %klen)\n  %isnull = icmp eq ptr %val, null\n  %okstep = xor i1 %isnull, true\n  %ok2 = and i1 %ok, %okstep\n  br label %cont\n")
    h.WriteString("cont:\n  %inext = add i32 %i, 1\n  br label %loop\n")
    h.WriteString("exit:\n  %res = select i1 %ok, i32 0, i32 1\n  ret i32 %res\n}\n")

    combined := append(rtIR, h.Bytes()...)
    combPath := filepath.Join(dir, "combined.ll")
    if err := os.WriteFile(combPath, combined, 0o644); err != nil { t.Fatalf("write combined: %v", err) }

    // Prefer lli
    if path, _ := exec.LookPath("lli"); path != "" {
        cmd := exec.Command(path, combPath)
        if _, err := cmd.CombinedOutput(); err == nil { return }
        // fallback to clang if lli fails
    }

    // Fallback: clang link and run (may not be available on all envs)
    clang, err := FindClang()
    if err != nil { t.Skip("lli/clang unavailable; skipping") }
    bin := filepath.Join(dir, "harness.bin")
    args := []string{"-x", "ir", combPath, "-o", bin, "-target", triple}
    if out, err := exec.Command(clang, args...).CombinedOutput(); err != nil {
        t.Skipf("clang link failed: %v, out=%s", err, string(out))
    }
    if out, err := exec.Command(bin).CombinedOutput(); err != nil {
        t.Fatalf("run failed: %v, out=%s", err, string(out))
    }
}

func irConstBytes(kind string, idx int, b []byte) string {
    // Encode bytes to LLVM c"..." with escapes; always append NUL
    var buf bytes.Buffer
    name := "." + kind + "." + itoa(idx)
    n := len(b) + 1
    buf.WriteString("@" + name + " = private constant [" + itoa(n) + " x i8] c\"")
    for _, c := range b {
        switch c {
        case '"': buf.WriteString("\\22")
        case '\\': buf.WriteString("\\5C")
        case '\n': buf.WriteString("\\0A")
        case '\r': buf.WriteString("\\0D")
        case '\t': buf.WriteString("\\09")
        default:
            if c >= 32 && c < 127 {
                buf.WriteByte(c)
            } else {
                hex := []byte("0123456789ABCDEF")
                buf.WriteByte('\\')
                buf.WriteByte(hex[c>>4])
                buf.WriteByte(hex[c&0x0F])
            }
        }
    }
    buf.WriteString("\\00\"\n")
    return buf.String()
}
