package driver

import (
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "github.com/sam-caldwell/ami/src/ami/compiler/sem"
)

// lowerFile lowers functions found in an AST file into a single IR module.
func lowerFile(pkg string, f *ast.File, params map[string][]string, results map[string][]string, paramNames map[string][]string) ir.Module {
    // signature maps are provided by caller (compile phase)
    var fns []ir.Function
    var concurrency int
    var backpressure string
    var telemetry bool
    var schedule string
    var capabilities []string
    var trustLevel string
    // Compute SCCs for recursion analysis
    scc := sem.ComputeSCC(f)
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok {
            var same map[string]bool
            if set, ok := scc[fn.Name]; ok { same = set }
            fns = append(fns, lowerFuncDeclWithSCC(fn, results, params, paramNames, same))
        }
    }
    // collect directives from pragmas
    var dirs []ir.Directive
    if f != nil {
        for _, pr := range f.Pragmas {
            switch pr.Domain {
            case "concurrency":
                // level parameter
                if lv, ok := pr.Params["level"]; ok {
                    // naive atoi
                    n := 0
                    for i := 0; i < len(lv); i++ { if lv[i] >= '0' && lv[i] <= '9' { n = n*10 + int(lv[i]-'0') } else { n = 0; break } }
                    if n > 0 { concurrency = n }
                }
                // schedule may be provided as a keyed pragma: `#pragma concurrency:schedule fair`
                if pr.Key == "schedule" {
                    if pr.Value != "" { schedule = strings.ToLower(pr.Value) }
                }
                // or as a param e.g., schedule=fair; also accept policy alias
                if v, ok := pr.Params["schedule"]; ok && schedule == "" { schedule = strings.ToLower(v) }
                if v, ok := pr.Params["policy"]; ok && schedule == "" { schedule = strings.ToLower(v) }
            case "backpressure":
                if pol, ok := pr.Params["policy"]; ok { backpressure = pol }
            case "telemetry":
                if en, ok := pr.Params["enabled"]; ok && (en == "true" || en == "1") { telemetry = true }
            case "schedule":
                // legacy/alternate domain
                if pr.Value != "" { schedule = strings.ToLower(pr.Value) }
                if v, ok := pr.Params["policy"]; ok && schedule == "" { schedule = strings.ToLower(v) }
            case "capabilities":
                // support list param as comma separated; also consider args list
                if lst, ok := pr.Params["list"]; ok && lst != "" {
                    for _, p := range strings.Split(lst, ",") {
                        p = strings.TrimSpace(p)
                        if p != "" { capabilities = append(capabilities, p) }
                    }
                }
                if len(pr.Args) > 0 { capabilities = append(capabilities, pr.Args...) }
            case "trust":
                if lv, ok := pr.Params["level"]; ok { trustLevel = lv }
            }
            dirs = append(dirs, ir.Directive{Domain: pr.Domain, Key: pr.Key, Value: pr.Value, Args: append([]string(nil), pr.Args...), Params: pr.Params})
        }
    }
    // derive capabilities from io.* and net.* usage in pipelines (ingress/egress allowed)
    if f != nil {
        addCap := func(cap string) {
            if cap == "" { return }
            for _, c := range capabilities { if c == cap { return } }
            capabilities = append(capabilities, cap)
        }
        for _, d := range f.Decls {
            pd, ok := d.(*ast.PipelineDecl)
            if !ok { continue }
            var steps []*ast.StepStmt
            for _, s := range pd.Stmts { if st, ok := s.(*ast.StepStmt); ok { steps = append(steps, st) } }
            if len(steps) == 0 { continue }
            for _, st := range steps {
                if strings.HasPrefix(st.Name, "io.") {
                    if cap := mapIOCapability(st.Name); cap != "" { addCap(cap) }
                }
                if strings.HasPrefix(strings.ToLower(st.Name), "net.") {
                    addCap("net")
                }
            }
        }
    }
    // Capability normalization: include generic 'io' when specifics are present
    hasIO := false
    hasIOSpecific := false
    for _, c := range capabilities { if c == "io" { hasIO = true } }
    for _, c := range capabilities { if len(c) > 3 && c[:3] == "io." { hasIOSpecific = true; break } }
    if hasIOSpecific && !hasIO { capabilities = append(capabilities, "io") }
    // capture error pipelines (for backend metadata embedding)
    var errpipes []ir.ErrorPipeline
    if f != nil {
        for _, d := range f.Decls {
            pd, ok := d.(*ast.PipelineDecl)
            if !ok || pd.Error == nil || pd.Error.Body == nil { continue }
            var names []string
            for _, s := range pd.Error.Body.Stmts {
                if st, ok := s.(*ast.StepStmt); ok {
                    names = append(names, st.Name)
                }
            }
            if len(names) > 0 {
                errpipes = append(errpipes, ir.ErrorPipeline{Pipeline: pd.Name, Steps: names})
            }
        }
    }
    // standard event lifecycle metadata descriptor for observability
    em := &ir.EventMeta{Schema: "eventmeta.v1", Fields: []string{"id", "ts", "attempt", "trace"}}
    // extract pipeline merge/collect specs into IR for downstream runtimes
    pipes := lowerPipelines(f)
    return ir.Module{Package: pkg, Functions: fns, Directives: dirs, Pipelines: pipes, ErrorPipes: errpipes, Concurrency: concurrency, Backpressure: backpressure, Schedule: schedule, TelemetryEnabled: telemetry, Capabilities: capabilities, TrustLevel: trustLevel, EventMeta: em}
}

// mapIOCapability converts an io.* step name into a coarse-grained capability string.
// mapIOCapability moved to lower_file_mapio.go to satisfy single-declaration rule
