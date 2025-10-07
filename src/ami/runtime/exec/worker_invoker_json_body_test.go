package exec

import (
    "context"
    "encoding/json"
    "os/exec"
    "os"
    "path/filepath"
    "runtime"
    "testing"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// compileJSONWorkers builds a shared library exporting two workers using the JSON ABI:
//  - ami_worker_EvJson: returns a full Event JSON with payload {"x":7}
//  - ami_worker_PayloadJson: returns a bare JSON payload 42
func compileJSONWorkers(t *testing.T) string {
    t.Helper()
    dir := t.TempDir()
    cfile := filepath.Join(dir, "w.c")
    src := `
#include <stdlib.h>
#include <string.h>

// JSON ABI: const char* f(const char* in, int in_len, int* out_len, const char** err)

static char* dupstr(const char* s){ size_t n=strlen(s); char* b=(char*)malloc(n); if(!b) return NULL; memcpy(b,s,n); return b; }

const char* ami_worker_EvJson(const char* in, int in_len, int* out_len, const char** err){
  (void)in; (void)in_len; if(err) *err=NULL; const char* body = "{\"schema\":\"events.v1\",\"payload\":{\"x\":7}}"; size_t n=strlen(body);
  char* buf=(char*)malloc(n); if(!buf){ if(err) *err=dupstr("oom"); return NULL; } memcpy(buf, body, n); if(out_len) *out_len=(int)n; return buf;
}

const char* ami_worker_PayloadJson(const char* in, int in_len, int* out_len, const char** err){
  (void)in; (void)in_len; if(err) *err=NULL; const char* body = "42"; size_t n=strlen(body);
  char* buf=(char*)malloc(n); if(!buf){ if(err) *err=dupstr("oom"); return NULL; } memcpy(buf, body, n); if(out_len) *out_len=(int)n; return buf;
}
`
    if err := os.WriteFile(cfile, []byte(src), 0o644); err != nil { t.Fatal(err) }
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
    if out, err := cmd.CombinedOutput(); err != nil { t.Skipf("compiler unavailable or failed: %v, out=%s", err, string(out)) }
    return outLib
}

func TestWorkerInvoker_JSONBodies_EventAndPayload(t *testing.T) {
    // Arrange pipelines and edges
    pkg, pipe := "app", "P"
    // ingress -> Transform -> egress
    m := MakeModuleWithEdges(t, pkg, pipe, []edgeEntry{{Unit: pipe, Pipeline: pipe, From: "ingress", To: "Transform"}, {Unit: pipe, Pipeline: pipe, From: "Transform", To: "egress"}})
    // pipelines JSON with two runs: one using EvJson, another PayloadJson
    writePipe := func(worker string) {
        type pl struct{ Pipelines []struct{ Name string; Steps []struct{ Name string; Args []string } } }
        var obj pl
        obj.Pipelines = append(obj.Pipelines, struct{ Name string; Steps []struct{ Name string; Args []string } }{Name: pipe, Steps: []struct{ Name string; Args []string }{{Name: "ingress"}, {Name: "Transform", Args: []string{worker}}, {Name: "egress"}}})
        b, _ := json.MarshalIndent(obj, "", "  ")
        dirIR := filepath.Join("build", "debug", "ir", pkg)
        _ = os.MkdirAll(dirIR, 0o755)
        if err := os.WriteFile(filepath.Join(dirIR, "u.pipelines.json"), b, 0o644); err != nil { t.Fatal(err) }
    }
    lib := compileJSONWorkers(t)
    eng := &Engine{}
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Case 1: Event JSON body
    writePipe("EvJson")
    in := make(chan ev.Event, 1)
    in <- ev.Event{Payload: map[string]any{"i": 1}}
    close(in)
    opts := ExecOptions{Invoker: NewDLSOInvoker(lib, "ami_worker_")}
    out, statsCh, err := eng.RunPipelineWithStats(ctx, m, pipe, in, nil, "", "", opts)
    if err != nil { t.Fatalf("run: %v", err) }
    var gotEv []int
    for e := range out {
        if m, ok := e.Payload.(map[string]any); ok {
            switch x := m["x"].(type) { case float64: gotEv = append(gotEv, int(x)); case int: gotEv = append(gotEv, x) }
        }
    }
    for range statsCh {}
    if len(gotEv) != 1 || gotEv[0] != 7 { t.Fatalf("unexpected event json payload: %v", gotEv) }

    // Case 2: Payload JSON body
    writePipe("PayloadJson")
    in2 := make(chan ev.Event, 1)
    in2 <- ev.Event{Payload: 1}
    close(in2)
    out2, statsCh2, err := eng.RunPipelineWithStats(ctx, m, pipe, in2, nil, "", "", opts)
    if err != nil { t.Fatalf("run2: %v", err) }
    var got []int
    for e := range out2 { switch x := e.Payload.(type) { case float64: got = append(got, int(x)); case int: got = append(got, x) } }
    for range statsCh2 {}
    if len(got) != 1 || got[0] != 42 { t.Fatalf("unexpected payload json result: %v", got) }
}
