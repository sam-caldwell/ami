//go:build cgo && (linux || darwin)

package exec

/*
#cgo linux LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>

typedef const char* (*ami_worker_fn)(const char*, int, int*, const char**);

static void* ami_dlopen(const char* path) {
    if (path && path[0]) {
        return dlopen(path, RTLD_NOW|RTLD_LOCAL);
    }
    return RTLD_DEFAULT;
}

static void* ami_dlsym(void* h, const char* name) {
    return dlsym(h, name);
}

static const char* ami_call_worker(void* fn, const char* in, int in_len, int* out_len, const char** err) {
    ami_worker_fn f = (ami_worker_fn)fn;
    return f(in, in_len, out_len, err);
}

static void ami_dlclose(void* h) {
    if (h && h != RTLD_DEFAULT) dlclose(h);
}
*/
import "C"

import (
    "encoding/json"
    "errors"
    "unsafe"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// DLSOInvoker resolves worker symbols from a shared library or the current process
// using POSIX dlsym(). It expects symbol names following a prefix (e.g., "ami_worker_").
// The symbol must implement the minimal ABI:
//   const char* fn(const char* in_json, int in_len, int* out_len, const char** err);
// On success, returns malloc'd JSON (Event or payload) and sets *err to NULL. On error,
// returns NULL and sets *err to malloc'd error string. Caller frees returned pointers.
type DLSOInvoker struct {
    libPath string
    prefix  string
    handle  unsafe.Pointer
}

func (d *DLSOInvoker) open() {
    if d.handle != nil { return }
    cpath := C.CString(d.libPath)
    defer C.free(unsafe.Pointer(cpath))
    d.handle = C.ami_dlopen(cpath)
}

// Close releases the library handle if applicable.
func (d *DLSOInvoker) Close() { if d.handle != nil { C.ami_dlclose(d.handle); d.handle = nil } }

// Resolve looks up a worker symbol and returns a Go wrapper when found.
func (d *DLSOInvoker) Resolve(workerName string) (func(ev.Event) (any, error), bool) {
    d.open()
    if d.handle == nil { return nil, false }
    symName := SanitizeWorkerSymbol(d.prefix, workerName)
    csym := C.CString(symName)
    defer C.free(unsafe.Pointer(csym))
    fn := C.ami_dlsym(d.handle, csym)
    if fn == nil { return nil, false }
    call := func(e ev.Event) (any, error) {
        inb, _ := json.Marshal(e)
        if len(inb) == 0 { inb = []byte("{}") }
        var outLen C.int
        var errStr *C.char
        cin := (*C.char)(unsafe.Pointer(&inb[0]))
        cout := C.ami_call_worker(fn, cin, C.int(len(inb)), &outLen, (**C.char)(unsafe.Pointer(&errStr)))
        if cout == nil {
            if errStr != nil { defer C.free(unsafe.Pointer(errStr)); return nil, errors.New(C.GoString(errStr)) }
            return nil, errors.New("worker returned null output")
        }
        defer C.free(unsafe.Pointer(cout))
        outBytes := C.GoBytes(unsafe.Pointer(cout), outLen)
        // Try to interpret as Event first
        var evOut ev.Event
        if err := json.Unmarshal(outBytes, &evOut); err == nil && (evOut.Payload != nil || len(outBytes) > 0) {
            return evOut, nil
        }
        // Fallback to bare payload
        var payload any
        if err := json.Unmarshal(outBytes, &payload); err == nil {
            return payload, nil
        }
        return nil, errors.New("invalid worker output JSON")
    }
    return call, true
}
