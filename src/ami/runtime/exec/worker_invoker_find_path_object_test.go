//go:build cgo && linux

package exec

import (
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "testing"

    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// Build a shared library with C workers that exercise ami_rt_find_path over objects
// including escaped quotes, unicode \uXXXX in keys, nested arrays/objects, and comma/whitespace edges.
func TestWorkerInvoker_FindPath_ObjectKeys(t *testing.T) {
    clang, err := llvme.FindClang(); if err != nil { t.Skip("clang not found") }
    if ver, err := llvme.Version(clang); err == nil && ver == "" { t.Skip("clang version undetectable; skipping") }
    triple := llvme.DefaultTriple
    dir := t.TempDir()

    // Runtime
    rtDir := filepath.Join(dir, "rt")
    if _, err := llvme.WriteRuntimeLL(rtDir, triple, false); err != nil { t.Fatalf("WriteRuntimeLL: %v", err) }
    rtLL := filepath.Join(rtDir, "runtime.ll")
    rtObj := filepath.Join(rtDir, "runtime.o")
    if err := llvme.CompileLLToObject(clang, rtLL, rtObj, triple); err != nil { t.Skipf("compile runtime.ll failed: %v", err) }

    // C workers: each returns a numeric JSON encoded result from ami_rt_event_get_i64
    cSrc := `
#include <stdlib.h>
#include <string.h>

extern void* ami_rt_json_to_event(const char*, int);
extern long long ami_rt_event_get_i64(void*, const char*, int);
extern const char* ami_rt_i64_to_json(long long, int*);

static const char* make_json_i64(long long v, int* out_len){ return ami_rt_i64_to_json(v, out_len); }

// 1) Escaped quotes in key: {"a\"b": 42}
const char* ami_worker_KEsc(const char* in_json, int in_len, int* out_len, const char** err){ (void)in_json; (void)in_len; if(err) *err=NULL; const char* js = "{\"a\\\"b\": 42}"; void* ev = ami_rt_json_to_event(js, (int)strlen(js)); long long n = ami_rt_event_get_i64(ev, "a\"b", 3); return make_json_i64(n, out_len); }

// 2) Unicode escape in key: {"A\u0022B": 99} where \u0022 -> '"'
const char* ami_worker_KUnicode(const char* in_json, int in_len, int* out_len, const char** err){ (void)in_json; (void)in_len; if(err) *err=NULL; const char* js = "{\"A\\u0022B\": 99}"; void* ev = ami_rt_json_to_event(js, (int)strlen(js)); long long n = ami_rt_event_get_i64(ev, "A\"B", 3); return make_json_i64(n, out_len); }

// 3) Nested object/array with key containing quote inside nested object: {"nested":{"x":[0,{"a\"b":7}]}}
const char* ami_worker_KNested(const char* in_json, int in_len, int* out_len, const char** err){ (void)in_json; (void)in_len; if(err) *err=NULL; const char* js = "{\"nested\":{\"x\":[0,{\"a\\\"b\":7}]}}"; void* ev = ami_rt_json_to_event(js, (int)strlen(js)); const char* path = "nested.x.1.a\"b"; long long n = ami_rt_event_get_i64(ev, path, (int)strlen(path)); return make_json_i64(n, out_len); }

// 4) Comma/whitespace edges: {"k"  :    1 , "m":2}
const char* ami_worker_KWS(const char* in_json, int in_len, int* out_len, const char** err){ (void)in_json; (void)in_len; if(err) *err=NULL; const char* js = "{\"k\"  :    1 , \"m\":2}"; void* ev = ami_rt_json_to_event(js, (int)strlen(js)); long long n = ami_rt_event_get_i64(ev, "k", 1); return make_json_i64(n, out_len); }

// 5) Ensure no false positive from string value containing key text: {"s":"a\"b"}
const char* ami_worker_KNoFP(const char* in_json, int in_len, int* out_len, const char** err){ (void)in_json; (void)in_len; if(err) *err=NULL; const char* js = "{\"s\":\"a\\\"b\"}"; void* ev = ami_rt_json_to_event(js, (int)strlen(js)); long long n = ami_rt_event_get_i64(ev, "a\"b", 3); return make_json_i64(n, out_len); }

// 6) Backslash in key: {"a\\b":55}
const char* ami_worker_KBackslash(const char* in_json, int in_len, int* out_len, const char** err){ (void)in_json; (void)in_len; if(err) *err=NULL; const char* js = "{\"a\\\\b\":55}"; void* ev = ami_rt_json_to_event(js, (int)strlen(js)); long long n = ami_rt_event_get_i64(ev, "a\\b", 3); return make_json_i64(n, out_len); }

    // 7) Multiple adjacent escapes: key q\u0022\u0041\"z -> q"Az
    const char* ami_worker_KMultiEsc(const char* in_json, int in_len, int* out_len, const char** err){ (void)in_json; (void)in_len; if(err) *err=NULL; const char* js = "{\"q\\u0022\\u0041\\\"z\":123}"; void* ev = ami_rt_json_to_event(js, (int)strlen(js)); long long n = ami_rt_event_get_i64(ev, "q\"Az", 4); return make_json_i64(n, out_len); }

// 8) Emoji key via surrogate pair: {"\uD83D\uDE00":321} where U+1F600 (😀), path is UTF-8 bytes
const char* ami_worker_KEmoji(const char* in_json, int in_len, int* out_len, const char** err){ (void)in_json; (void)in_len; if(err) *err=NULL; const char* js = "{\"\\uD83D\\uDE00\":321}"; void* ev = ami_rt_json_to_event(js, (int)strlen(js)); const char* p = "\xF0\x9F\x98\x80"; long long n = ami_rt_event_get_i64(ev, p, 4); return make_json_i64(n, out_len); }

// 9) 2-byte UTF-8: {"\u00E9": 77} (é -> C3 A9)
const char* ami_worker_K2Byte(const char* in_json, int in_len, int* out_len, const char** err){ (void)in_json; (void)in_len; if(err) *err=NULL; const char* js = "{\"\\u00E9\":77}"; void* ev = ami_rt_json_to_event(js, (int)strlen(js)); const char* p = "\xC3\xA9"; long long n = ami_rt_event_get_i64(ev, p, 2); return make_json_i64(n, out_len); }

// 10) 3-byte UTF-8: {"\u20AC": 88} (Euro sign -> E2 82 AC)
const char* ami_worker_K3Byte(const char* in_json, int in_len, int* out_len, const char** err){ (void)in_json; (void)in_len; if(err) *err=NULL; const char* js = "{\"\\u20AC\":88}"; void* ev = ami_rt_json_to_event(js, (int)strlen(js)); const char* p = "\xE2\x82\xAC"; long long n = ami_rt_event_get_i64(ev, p, 3); return make_json_i64(n, out_len); }
`
    cPath := filepath.Join(dir, "k.c")
    if err := os.WriteFile(cPath, []byte(cSrc), 0o644); err != nil { t.Fatal(err) }
    cObj := filepath.Join(dir, "k.o")
    if out, err := exec.Command(clang, "-c", cPath, "-o", cObj, "-target", triple).CombinedOutput(); err != nil { t.Skipf("compile k.c failed: %v, out=%s", err, string(out)) }

    // Link shared library with runtime
    var lib string
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "linux":
        lib = filepath.Join(dir, "libk.so")
        cmd = exec.Command(clang, "-shared", "-fPIC", cObj, rtObj, "-o", lib, "-target", triple)
    case "darwin":
        lib = filepath.Join(dir, "libk.dylib")
        cmd = exec.Command(clang, "-dynamiclib", cObj, rtObj, "-o", lib, "-target", triple)
    default:
        t.Skip("OS not supported for shared linking")
    }
    if out, err := cmd.CombinedOutput(); err != nil { t.Skipf("link shared lib failed: %v, out=%s", err, string(out)) }

    inv := NewDLSOInvoker(lib, "ami_worker_")
    if inv == nil { t.Skip("invoker unavailable (no cgo)") }
    inEv := ev.Event{Payload: 1}

    // KEsc -> 42
    if f, ok := inv.Resolve("KEsc"); !ok || f == nil { t.Fatalf("resolve KEsc failed") } else {
        v, err := f(inEv)
        if err != nil { t.Fatalf("KEsc: %v", err) }
        switch x := v.(type) {
        case float64:
            if int(x) != 42 { t.Fatalf("KEsc got %v", v) }
        default:
            t.Fatalf("KEsc type %T", v)
        }
    }
    // KUnicode -> 99
    if f, ok := inv.Resolve("KUnicode"); !ok || f == nil { t.Fatalf("resolve KUnicode failed") } else {
        v, err := f(inEv)
        if err != nil { t.Fatalf("KUnicode: %v", err) }
        switch x := v.(type) {
        case float64:
            if int(x) != 99 { t.Fatalf("KUnicode got %v", v) }
        default:
            t.Fatalf("KUnicode type %T", v)
        }
    }
    // KNested -> 7
    if f, ok := inv.Resolve("KNested"); !ok || f == nil { t.Fatalf("resolve KNested failed") } else {
        v, err := f(inEv)
        if err != nil { t.Fatalf("KNested: %v", err) }
        switch x := v.(type) {
        case float64:
            if int(x) != 7 { t.Fatalf("KNested got %v", v) }
        default:
            t.Fatalf("KNested type %T", v)
        }
    }
    // KWS -> 1
    if f, ok := inv.Resolve("KWS"); !ok || f == nil { t.Fatalf("resolve KWS failed") } else {
        v, err := f(inEv)
        if err != nil { t.Fatalf("KWS: %v", err) }
        switch x := v.(type) {
        case float64:
            if int(x) != 1 { t.Fatalf("KWS got %v", v) }
        default:
            t.Fatalf("KWS type %T", v)
        }
    }
    // KNoFP -> expect 0 (not found)
    if f, ok := inv.Resolve("KNoFP"); !ok || f == nil { t.Fatalf("resolve KNoFP failed") } else {
        v, err := f(inEv)
        if err != nil { t.Fatalf("KNoFP: %v", err) }
        switch x := v.(type) {
        case float64:
            if int(x) != 0 { t.Fatalf("KNoFP got %v", v) }
        default:
            t.Fatalf("KNoFP type %T", v)
        }
    }

    // KBackslash -> 55
    if f, ok := inv.Resolve("KBackslash"); !ok || f == nil { t.Fatalf("resolve KBackslash failed") } else {
        v, err := f(inEv)
        if err != nil { t.Fatalf("KBackslash: %v", err) }
        switch x := v.(type) {
        case float64:
            if int(x) != 55 { t.Fatalf("KBackslash got %v", v) }
        default:
            t.Fatalf("KBackslash type %T", v)
        }
    }

    // KMultiEsc -> 123
    if f, ok := inv.Resolve("KMultiEsc"); !ok || f == nil { t.Fatalf("resolve KMultiEsc failed") } else {
        v, err := f(inEv)
        if err != nil { t.Fatalf("KMultiEsc: %v", err) }
        switch x := v.(type) {
        case float64:
            if int(x) != 123 { t.Fatalf("KMultiEsc got %v", v) }
        default:
            t.Fatalf("KMultiEsc type %T", v)
        }
    }

    // KEmoji -> 321
    if f, ok := inv.Resolve("KEmoji"); !ok || f == nil { t.Fatalf("resolve KEmoji failed") } else {
        v, err := f(inEv)
        if err != nil { t.Fatalf("KEmoji: %v", err) }
        switch x := v.(type) {
        case float64:
            if int(x) != 321 { t.Fatalf("KEmoji got %v", v) }
        default:
            t.Fatalf("KEmoji type %T", v)
        }
    }

    // K2Byte -> 77
    if f, ok := inv.Resolve("K2Byte"); !ok || f == nil { t.Fatalf("resolve K2Byte failed") } else {
        v, err := f(inEv)
        if err != nil { t.Fatalf("K2Byte: %v", err) }
        switch x := v.(type) {
        case float64:
            if int(x) != 77 { t.Fatalf("K2Byte got %v", v) }
        default:
            t.Fatalf("K2Byte type %T", v)
        }
    }

    // K3Byte -> 88
    if f, ok := inv.Resolve("K3Byte"); !ok || f == nil { t.Fatalf("resolve K3Byte failed") } else {
        v, err := f(inEv)
        if err != nil { t.Fatalf("K3Byte: %v", err) }
        switch x := v.(type) {
        case float64:
            if int(x) != 88 { t.Fatalf("K3Byte got %v", v) }
        default:
            t.Fatalf("K3Byte type %T", v)
        }
    }
}
