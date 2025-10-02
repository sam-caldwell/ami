package main

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"
    "strings"
    "strconv"

    "github.com/spf13/cobra"
    "github.com/sam-caldwell/ami/src/ami/compiler/driver"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    rexec "github.com/sam-caldwell/ami/src/ami/runtime/exec"
    rmerge "github.com/sam-caldwell/ami/src/ami/runtime/merge"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func newRunCmd() *cobra.Command {
    var pkgName string
    var pipeline string
    var eventsPath string
    var limit int
    var timeout string
    var stats bool
    var collectIndex int
    var srcType string
    var rate string
    var count int
    var sink string
    var sinkPath string
    var format string
    var filterExpr string
    var transformExpr string
    cmd := &cobra.Command{
        Use:   "run",
        Short: "Simulate a pipeline with merge Collect nodes using IR + runtime executor",
        RunE: func(cmd *cobra.Command, args []string) error {
            wd, err := os.Getwd(); if err != nil { return err }
            if pkgName == "" || pipeline == "" { return fmt.Errorf("--package and --pipeline are required") }
            // Load workspace
            var ws workspace.Workspace
            if err := ws.Load(filepath.Join(wd, "ami.workspace")); err != nil { return err }
            // Collect package files and compile to emit IR JSON
            var pkgs []driver.Package
            entry := ws.FindPackage(pkgName)
            if entry == nil { return fmt.Errorf("package not found: %s", pkgName) }
            root := filepath.Clean(filepath.Join(wd, entry.Root))
            var files []string
            _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error { if err == nil && !d.IsDir() && filepath.Ext(path) == ".ami" { files = append(files, path) }; return nil })
            if len(files) == 0 { return fmt.Errorf("no .ami files in %s", root) }
            var fs source.FileSet
            for _, f := range files { b, err := os.ReadFile(f); if err == nil { fs.AddFile(f, string(b)) } }
            pkgs = append(pkgs, driver.Package{Name: pkgName, Files: &fs})
            arts, _ := driver.Compile(ws, pkgs, driver.Options{Debug: true})
            if len(arts.IR) == 0 { return fmt.Errorf("no IR artifacts") }
            // Find module with target pipeline
            var module ir.Module
            found := false
            for _, path := range arts.IR {
                b, err := os.ReadFile(path); if err != nil { continue }
                var m ir.Module
                if err := json.Unmarshal(b, &m); err != nil { continue }
                for _, p := range m.Pipelines { if p.Name == pipeline { module = m; found = true; break } }
                if found { break }
            }
            if !found { return fmt.Errorf("pipeline not found: %s.%s", pkgName, pipeline) }
            // Build executor and channels
            eng, err := rexec.NewEngineFromModule(module); if err != nil { return err }
            defer eng.Close()
            in := make(chan ev.Event, 1024)
            // timeout / cancellation
            base := context.Background()
            ctx, cancel := context.WithCancel(base)
            defer cancel()
            if timeout != "" { if d, e := time.ParseDuration(timeout); e == nil { c, cancel2 := context.WithTimeout(ctx, d); ctx = c; defer cancel2() } }
            // decide single collect or full pipeline; wire stats emitter when requested
            var out <-chan ev.Event
            statsCh := make(chan map[string]any, 16)
            var stageStats <-chan rexec.StageStats
            emitStage := func(info rexec.StageInfo, st rmerge.Stats) {
                statsCh <- map[string]any{
                    "schema":  "stage.stats.v1",
                    "stage":   map[string]any{"name": info.Name, "kind": info.Kind, "index": info.Index},
                    "enqueued": st.Enqueued, "emitted": st.Emitted, "dropped": st.Dropped, "expired": st.Expired,
                }
            }
            if collectIndex >= 0 {
                // specific collect
                var mp *ir.MergePlan
                for _, p := range module.Pipelines { if p.Name == pipeline { if collectIndex < len(p.Collect) { if p.Collect[collectIndex].Merge != nil { mp = p.Collect[collectIndex].Merge } } } }
                if mp == nil { return fmt.Errorf("collect index %d not found for pipeline %s", collectIndex, pipeline) }
                ch := make(chan ev.Event, 1024)
                go func(prev <-chan ev.Event, next chan<- ev.Event){ for e := range prev { next <- e }; close(next) }(in, ch)
                oc, s, err := eng.RunMergeWithStats(ctx, *mp, ch); if err != nil { return err }
                if stats { go emitStage(rexec.StageInfo{Name: "Collect", Kind: "collect", Index: collectIndex}, *s) }
                out = oc
            } else {
                var oc <-chan ev.Event
                if stats {
                    outCh, sc, err := eng.RunPipelineWithStats(ctx, module, pipeline, in, emitStage, filterExpr, transformExpr, rexec.ExecOptions{SourceType: srcType, TimerInterval: parseRate(rate), TimerCount: count})
                    if err != nil { return err }
                    oc = outCh
                    stageStats = sc
                } else {
                    outCh, err := eng.RunPipeline(ctx, module, pipeline, in)
                    if err != nil { return err }
                    oc = outCh
                }
                out = oc
            }
            // Source: file/stdin or timer
            done := make(chan struct{})
            switch srcType {
            case "timer":
                d := 100 * time.Millisecond
                if rate != "" {
                    if strings.Contains(rate, "/s") {
                        // events per second
                        nstr := strings.TrimSuffix(rate, "/s")
                        if n, err := strconv.Atoi(nstr); err == nil && n > 0 { d = time.Second / time.Duration(n) }
                    } else if rdur, err := time.ParseDuration(rate); err == nil { d = rdur }
                }
                go func(){
                    max := count
                    for i := 0; max == 0 || i < max; i++ {
                        in <- ev.Event{Payload: map[string]any{"i": i, "ts": time.Now().UTC()}}
                        time.Sleep(d)
                    }
                    close(in); close(done)
                }()
            default:
                var r *bufio.Scanner
                if eventsPath == "" || eventsPath == "-" { r = bufio.NewScanner(os.Stdin) } else { f, err := os.Open(eventsPath); if err != nil { return err }; defer f.Close(); r = bufio.NewScanner(f) }
                go func(){
                    for r.Scan() {
                        line := r.Bytes()
                        var e ev.Event
                        if json.Unmarshal(line, &e) == nil && (e.Payload != nil || e.ID != "") { in <- e; continue }
                        var obj any
                        if json.Unmarshal(line, &obj) == nil { in <- ev.Event{Payload: obj}; continue }
                    }
                    time.Sleep(5 * time.Millisecond); close(in); close(done)
                }()
            }
            jsonOut, _ := cmd.Flags().GetBool("json")
            outWriter := cmd.OutOrStdout()
            var enc *json.Encoder
            var outFile *os.File
            if sink == "file" && sinkPath != "" {
                f, err := os.Create(sinkPath); if err != nil { return err }
                outFile = f; outWriter = f
            }
            if format == "pretty" { enc = json.NewEncoder(outWriter); enc.SetIndent("", "  ") } else { enc = json.NewEncoder(outWriter) }
            printed := 0
            start := time.Now()
            for e := range out {
                if format == "jsonl" || format == "pretty" { _ = enc.Encode(e) } else { _ = enc.Encode(e) }
                printed++
                if limit > 0 && printed >= limit { break }
            }
            // Drain and emit stage stats if flagged
            if stats {
                cancel()
                time.Sleep(10 * time.Millisecond)
                // Single-collect stats (map objects)
                for {
                    select {
                    case s := <-statsCh:
                        if s == nil { goto drainStages }
                        _ = enc.Encode(s)
                    default:
                        goto drainStages
                    }
                }
            drainStages:
                // Pipeline stage stats (typed channel)
                if stageStats != nil {
                    for s := range stageStats {
                        _ = enc.Encode(map[string]any{
                            "schema":  "stage.stats.v1",
                            "stage":   map[string]any{"name": s.Stage.Name, "kind": s.Stage.Kind, "index": s.Stage.Index},
                            "enqueued": s.Stats.Enqueued,
                            "emitted":  s.Stats.Emitted,
                            "dropped":  s.Stats.Dropped,
                            "expired":  s.Stats.Expired,
                        })
                    }
                }
            }
            if stats {
                dur := time.Since(start).Milliseconds()
                if jsonOut {
                    // Emit a final stats object for consumers
                    _ = enc.Encode(map[string]any{"schema":"run.stats.v1","outputs":printed,"duration_ms":dur})
                } else {
                    fmt.Fprintf(cmd.ErrOrStderr(), "outputs=%d duration_ms=%d\n", printed, dur)
                }
            }
            // Ensure input goroutine exits
            select { case <-done: default: }
            if outFile != nil { _ = outFile.Close() }
            return nil
        },
    }
    cmd.Flags().StringVar(&pkgName, "package", "", "package name containing the pipeline")
    cmd.Flags().StringVar(&pipeline, "pipeline", "", "pipeline name to run")
    cmd.Flags().StringVar(&eventsPath, "events", "-", "path to JSON Lines events (or '-' for stdin)")
    cmd.Flags().IntVar(&limit, "limit", 0, "stop after N outputs (0 = unlimited)")
    cmd.Flags().StringVar(&timeout, "timeout", "", "overall run timeout (e.g., 5s, 1m)")
    cmd.Flags().BoolVar(&stats, "stats", false, "print a summary of outputs and duration")
    cmd.Flags().IntVar(&collectIndex, "collect-index", -1, "run only the specified Collect step index (default: chain all)")
    cmd.Flags().StringVar(&srcType, "source", "auto", "event source: auto|file|timer")
    cmd.Flags().StringVar(&rate, "rate", "", "timer rate (e.g., 10/s or 100ms)")
    cmd.Flags().IntVar(&count, "count", 0, "timer events to emit (0=unlimited)")
    cmd.Flags().StringVar(&sink, "sink", "stdout", "sink: stdout|file")
    cmd.Flags().StringVar(&sinkPath, "sink-path", "", "sink file path when --sink=file")
    cmd.Flags().StringVar(&format, "format", "jsonl", "output format: jsonl|pretty")
    cmd.Flags().StringVar(&filterExpr, "filter", "none", "filter DSL stub (e.g., drop_even)")
    cmd.Flags().StringVar(&transformExpr, "transform", "none", "transform DSL stub (e.g., add_field:flag)")
    return cmd
}

// parseRate moved to run_cmd_parse_rate.go
