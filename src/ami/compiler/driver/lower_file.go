package driver

import (
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerFile lowers functions found in an AST file into a single IR module.
func lowerFile(pkg string, f *ast.File, params map[string][]string, results map[string][]string, paramNames map[string][]string) ir.Module {
    // signature maps are provided by caller (compile phase)
    var fns []ir.Function
    var concurrency int
    var backpressure string
    var telemetry bool
    var capabilities []string
    var trustLevel string
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok {
            fns = append(fns, lowerFuncDecl(fn, results, params, paramNames))
        }
    }
    // collect directives from pragmas
    var dirs []ir.Directive
    if f != nil {
        for _, pr := range f.Pragmas {
            switch pr.Domain {
            case "concurrency":
                if lv, ok := pr.Params["level"]; ok {
                    // naive atoi
                    n := 0
                    for i := 0; i < len(lv); i++ { if lv[i] >= '0' && lv[i] <= '9' { n = n*10 + int(lv[i]-'0') } else { n = 0; break } }
                    if n > 0 { concurrency = n }
                }
            case "backpressure":
                if pol, ok := pr.Params["policy"]; ok { backpressure = pol }
            case "telemetry":
                if en, ok := pr.Params["enabled"]; ok && (en == "true" || en == "1") { telemetry = true }
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
    // derive capabilities from io.* usage in pipelines (ingress/egress allowed)
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
            }
        }
    }
    return ir.Module{Package: pkg, Functions: fns, Directives: dirs, Concurrency: concurrency, Backpressure: backpressure, TelemetryEnabled: telemetry, Capabilities: capabilities, TrustLevel: trustLevel}
}

// mapIOCapability converts an io.* step name into a coarse-grained capability string.
func mapIOCapability(name string) string {
    // name like "io.Read", "io.WriteFile", "io.Open", "io.Connect", etc.
    s := strings.TrimPrefix(name, "io.")
    if s == name || s == "" { return "io" }
    // take leading identifier portion
    end := len(s)
    for i := 0; i < len(s); i++ {
        if s[i] < 'A' || (s[i] > 'Z' && s[i] < 'a') || s[i] > 'z' { end = i; break }
    }
    head := s[:end]
    switch strings.ToLower(head) {
    case "read", "readfile", "recv":
        return "io.read"
    case "write", "writefile", "send":
        return "io.write"
    case "open", "close":
        return "io.fs"
    case "connect", "listen", "dial":
        return "network"
    default:
        return "io"
    }
}
