package exec

import (
    "context"
    "encoding/json"
    "os/exec"
    "runtime"
    "os"
    "path/filepath"
    "testing"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
    errs "github.com/sam-caldwell/ami/src/schemas/errors"
)

// compileSharedLib writes src to path and compiles a shared library providing
// worker symbols using the minimal ABI required by DLSOInvoker.
func compileSharedLib(t *testing.T, src string) string {
    t.Helper()
    dir := t.TempDir()
    cfile := filepath.Join(dir, "w.c")
    if err := os.WriteFile(cfile, []byte(src), 0o644); err != nil { t.Fatalf("write c: %v", err) }
    var outLib string
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "linux":
        outLib = filepath.Join(dir, "libw.so")
        cmd = exec.Command("cc", "-shared", "-fPIC", cfile, "-o", outLib)
    case "darwin":
        outLib = filepath.Join(dir, "libw.dylib")
        cmd = exec.Command("cc", "-dynamiclib", cfile, "-o", outLib)
    default:
        t.Skip("dynamic linking test not supported on this OS")
    }
    if out, err := cmd.CombinedOutput(); err != nil {
        t.Skipf("compiler unavailable or failed: %v, out=%s", err, string(out))
    }
    return outLib
}

func TestWorkerInvoker_DLSym_PayloadAndEvent(t *testing.T) {
    pkg := "app"
    pipeline := "P"
    // edges: ingress -> Transform -> egress
    m := MakeModuleWithEdges(t, pkg, pipeline, []edgeEntry{{Unit: pipeline, Pipeline: pipeline, From: "ingress", To: "Transform"}, {Unit: pipeline, Pipeline: pipeline, From: "Transform", To: "egress"}})
    // pipelines: Transform with worker=W (payload return)
    writePipelines := func(worker string) {
        type pipeList struct{ Pipelines []struct{ Name string; Steps []struct{ Name string; Args []string } } }
        var pl pipeList
        pl.Pipelines = append(pl.Pipelines, struct{ Name string; Steps []struct{ Name string; Args []string } }{Name: pipeline, Steps: []struct{ Name string; Args []string }{{Name: "ingress"}, {Name: "Transform", Args: []string{worker}}, {Name: "egress"}}})
        pb, _ := json.MarshalIndent(pl, "", "  ")
        dirIR := "build/debug/ir/" + pkg
        _ = writeFile(t, dirIR+"/u.pipelines.json", pb)
    }

    // Case 1: payload return via W
    writePipelines("W")
    // Compile shared lib with three worker symbols: W, Ev, Fail
    lib := compileSharedLib(t, `
#include <stdlib.h>
#include <string.h>

const char* ami_worker_W(const char* in_json, int in_len, int* out_len, const char** err) {
    (void)in_json; (void)in_len; if (err) *err = NULL;
    const char* payload = "42"; // bare payload JSON (number)
    char* buf = (char*)malloc(2);
    if (!buf) { if (err) *err = strdup("oom"); return NULL; }
    memcpy(buf, payload, 2);
    *out_len = 2;
    return (const char*)buf;
}

const char* ami_worker_Ev(const char* in_json, int in_len, int* out_len, const char** err) {
    (void)in_json; (void)in_len; if (err) *err = NULL;
    const char* body = "{\"schema\":\"events.v1\",\"payload\":77}";
    size_t n = strlen(body);
    char* buf = (char*)malloc(n);
    if (!buf) { if (err) *err = strdup("oom"); return NULL; }
    memcpy(buf, body, n);
    *out_len = (int)n;
    return (const char*)buf;
}

const char* ami_worker_Fail(const char* in_json, int in_len, int* out_len, const char** err) {
    (void)in_json; (void)in_len; (void)out_len;
    if (err) *err = strdup("synthetic failure");
    return NULL;
}
`)
    eng := &Engine{}
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    in := make(chan ev.Event, 1)
    in <- ev.Event{Payload: 41}
    close(in)
    opts := ExecOptions{Invoker: NewDLSOInvoker(lib, "ami_worker_")}
    out, statsCh, err := eng.RunPipelineWithStats(ctx, m, pipeline, in, nil, "", "", opts)
    if err != nil { t.Fatalf("run: %v", err) }
    var got []int
    for e := range out {
        switch x := e.Payload.(type) { case int: got = append(got, x); case float64: got = append(got, int(x)) }
    }
    for range statsCh {}
    if len(got) != 1 || got[0] != 42 { t.Fatalf("unexpected outputs (W): %v", got) }

    // Case 2: event return via Ev
    writePipelines("Ev")
    in2 := make(chan ev.Event, 1)
    in2 <- ev.Event{Payload: 1}
    close(in2)
    out2, statsCh2, err := eng.RunPipelineWithStats(ctx, m, pipeline, in2, nil, "", "", opts)
    if err != nil { t.Fatalf("run2: %v", err) }
    got = got[:0]
    for e := range out2 {
        switch x := e.Payload.(type) { case int: got = append(got, x); case float64: got = append(got, int(x)) }
    }
    for range statsCh2 {}
    if len(got) != 1 || got[0] != 77 { t.Fatalf("unexpected outputs (Ev): %v", got) }
}

func TestWorkerInvoker_DLSym_ErrorRouting(t *testing.T) {
    pkg := "app"
    pipeline := "P"
    m := MakeModuleWithEdges(t, pkg, pipeline, []edgeEntry{{Unit: pipeline, Pipeline: pipeline, From: "ingress", To: "Transform"}, {Unit: pipeline, Pipeline: pipeline, From: "Transform", To: "egress"}})
    // pipelines with failing worker
    type pipeList struct{ Pipelines []struct{ Name string; Steps []struct{ Name string; Args []string } } }
    var pl pipeList
    pl.Pipelines = append(pl.Pipelines, struct{ Name string; Steps []struct{ Name string; Args []string } }{Name: pipeline, Steps: []struct{ Name string; Args []string }{{Name: "ingress"}, {Name: "Transform", Args: []string{"Fail"}}, {Name: "egress"}}})
    pb, _ := json.MarshalIndent(pl, "", "  ")
    _ = writeFile(t, "build/debug/ir/"+pkg+"/u.pipelines.json", pb)

    eng := &Engine{}
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    in := make(chan ev.Event, 1)
    in <- ev.Event{Payload: 1}
    close(in)
    errCh := make(chan errs.Error, 1)
    // reuse compiled lib
    opts := ExecOptions{Invoker: NewDLSOInvoker(compileSharedLib(t, `
#include <stdlib.h>
#include <string.h>
const char* ami_worker_Fail(const char* in_json, int in_len, int* out_len, const char** err) {
    (void)in_json; (void)in_len; (void)out_len;
    if (err) *err = strdup("synthetic failure");
    return NULL;
}
`), "ami_worker_"), ErrorChan: errCh}
    out, statsCh, err := eng.RunPipelineWithStats(ctx, m, pipeline, in, nil, "", "", opts)
    if err != nil { t.Fatalf("run: %v", err) }
    // Expect no outputs, but one error on errCh
    var outs int
    for range out { outs++ }
    for range statsCh {}
    if outs != 0 { t.Fatalf("unexpected outputs: %d", outs) }
    select {
    case e := <-errCh:
        if e.Code == "" { t.Fatalf("missing error code in worker error: %+v", e) }
    default:
        t.Fatalf("expected worker error on ErrorChan")
    }
}

// writeFile writes a file ensuring its directory exists.
func writeFile(t *testing.T, path string, b []byte) error {
    t.Helper()
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    return os.WriteFile(path, b, 0o644)
}
