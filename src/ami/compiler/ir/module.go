package ir

import (
    "sort"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    edg "github.com/sam-caldwell/ami/src/ami/compiler/edge"
    sch "github.com/sam-caldwell/ami/src/schemas"
    "strconv"
    "strings"
)

type Function struct { Name string }

type Module struct {
    Package   string
    Unit      string // file path
    Functions []Function
    // Directive-derived attributes (scaffold)
    Concurrency int
    Capabilities []string
    Trust        string
    Backpressure string
    Pipelines    []PipelineIR
}

// FromASTFile builds a simple IR module enumerating function declarations.
func FromASTFile(pkg, unit string, f *astpkg.File) Module {
    m := Module{Package: pkg, Unit: unit}
    for _, d := range f.Decls {
        if fd, ok := d.(astpkg.FuncDecl); ok {
            m.Functions = append(m.Functions, Function{Name: fd.Name})
        }
    }
    sort.Slice(m.Functions, func(i, j int) bool { return m.Functions[i].Name < m.Functions[j].Name })
    return m
}

// ApplyDirectives sets module attributes derived from top-level pragmas.
func (m *Module) ApplyDirectives(dirs []astpkg.Directive) {
    for _, d := range dirs {
        switch strings.ToLower(d.Name) {
        case "concurrency":
            // naive parse int
            if n, err := strconv.Atoi(strings.Fields(d.Payload)[0]); err == nil { m.Concurrency = n }
        case "capabilities":
            m.Capabilities = splitCSV(d.Payload)
        case "trust":
            m.Trust = strings.TrimSpace(d.Payload)
        case "backpressure":
            m.Backpressure = strings.TrimSpace(d.Payload)
        }
    }
}

func splitCSV(s string) []string {
    var out []string
    for _, p := range strings.Split(s, ",") { v := strings.TrimSpace(p); if v != "" { out = append(out, v) } }
    return out
}

// ToSchema converts the module into a schemas.IRV1 for debug output.
func (m Module) ToSchema() sch.IRV1 {
    out := sch.IRV1{Schema: "ir.v1", Package: m.Package, File: m.Unit}
    for _, fn := range m.Functions {
        out.Functions = append(out.Functions, sch.IRFunction{Name: fn.Name, Blocks: []sch.IRBlock{{Label: "entry"}}})
    }
    return out
}

// PipelineIR captures lowered worker/factory references for deeper checks.
type PipelineIR struct {
    Name       string
    Steps      []StepIR
    ErrorSteps []StepIR
}

type StepIR struct {
    Node    string
    Workers []WorkerIR
    In      edg.Spec // optional: parsed from arg `in=edge.*(...)`
}

type WorkerIR struct {
    Name       string
    Kind       string // function|factory
    HasContext bool
    HasState   bool
    // Captured generic payload types from signature shapes.
    Input      string // Event<T> -> T
    OutputKind string // Event|Events|Error|Drop|Ack
    Output     string // Event<U>/Events<U>/Error<E> -> U/E
}

// LowerPipelines derives PipelineIR using parsed worker refs and function decl signatures.
func (m *Module) LowerPipelines(f *astpkg.File) {
    // build func map
    funs := map[string]astpkg.FuncDecl{}
    for _, d := range f.Decls {
        if fd, ok := d.(astpkg.FuncDecl); ok { funs[fd.Name] = fd }
    }
    // walk pipelines
    for _, d := range f.Decls {
        pd, ok := d.(astpkg.PipelineDecl)
        if !ok { continue }
        pir := PipelineIR{Name: pd.Name}
        mkWorkers := func(st astpkg.NodeCall) []WorkerIR {
            var out []WorkerIR
            for _, w := range st.Workers {
                wi := WorkerIR{Name: w.Name, Kind: w.Kind}
                if fd, ok := funs[w.Name]; ok {
                    if len(fd.Params) >= 3 {
                        p1 := fd.Params[0].Type
                        p3 := fd.Params[2].Type
                        wi.HasContext = (p1.Name == "Context")
                        wi.HasState = (p3.Name == "State" && p3.Ptr)
                        // capture generic payloads
                        // input is param[1] which should be Event<T>
                        p2 := fd.Params[1].Type
                        if p2.Name == "Event" && len(p2.Args) == 1 {
                            wi.Input = typeRefToString(p2.Args[0])
                        }
                        // output is single result; derive kind and payload
                        if len(fd.Result) == 1 {
                            r := fd.Result[0]
                            switch {
                            case r.Name == "Event" && len(r.Args) == 1 && !r.Slice:
                                wi.OutputKind = "Event"
                                wi.Output = typeRefToString(r.Args[0])
                            case r.Name == "Event" && len(r.Args) == 1 && r.Slice:
                                wi.OutputKind = "Events"
                                wi.Output = typeRefToString(r.Args[0])
                            case r.Name == "Error" && len(r.Args) == 1:
                                wi.OutputKind = "Error"
                                wi.Output = typeRefToString(r.Args[0])
                            case r.Name == "Drop" && len(r.Args) == 0:
                                wi.OutputKind = "Drop"
                            case r.Name == "Ack" && len(r.Args) == 0:
                                wi.OutputKind = "Ack"
                            }
                        }
                    }
                }
                out = append(out, wi)
            }
            return out
        }
        for _, st := range pd.Steps {
            step := StepIR{Node: st.Name, Workers: mkWorkers(st)}
            if spec, ok := parseEdgeSpecFromArgs(st.Args); ok { step.In = spec }
            pir.Steps = append(pir.Steps, step)
        }
        for _, st := range pd.ErrorSteps {
            step := StepIR{Node: st.Name, Workers: mkWorkers(st)}
            if spec, ok := parseEdgeSpecFromArgs(st.Args); ok { step.In = spec }
            pir.ErrorSteps = append(pir.ErrorSteps, step)
        }
        m.Pipelines = append(m.Pipelines, pir)
    }
}

// typeRefToString renders an AST TypeRef (without positions) into a concise
// string, including pointer, slice and generic arguments. Examples:
//   string, *Foo, []Bar, Event<T>, []Event<map<string,int>>
func typeRefToString(t astpkg.TypeRef) string {
    var b strings.Builder
    if t.Ptr { b.WriteByte('*') }
    if t.Slice { b.WriteString("[]") }
    b.WriteString(t.Name)
    if len(t.Args) > 0 {
        b.WriteByte('<')
        for i, a := range t.Args {
            if i > 0 { b.WriteByte(',') }
            b.WriteString(typeRefToString(a))
        }
        b.WriteByte('>')
    }
    return b.String()
}

// ToPipelinesSchema converts lowered Pipelines into the public PipelinesV1 schema.
func (m Module) ToPipelinesSchema() sch.PipelinesV1 {
    out := sch.PipelinesV1{Schema: "pipelines.v1", Package: m.Package, File: m.Unit}
    for _, p := range m.Pipelines {
        sp := sch.PipelineV1{Name: p.Name}
        // helper to convert steps
        conv := func(steps []StepIR) []sch.PipelineStepV1 {
            var res []sch.PipelineStepV1
            for _, st := range steps {
                ps := sch.PipelineStepV1{Node: st.Node}
                for _, w := range st.Workers {
                    ps.Workers = append(ps.Workers, sch.PipelineWorkerV1{
                        Name: w.Name, Kind: w.Kind, HasContext: w.HasContext, HasState: w.HasState,
                        Input: w.Input, OutputKind: w.OutputKind, Output: w.Output,
                    })
                }
                if st.In != nil {
                    ps.InEdge = toSchemaEdge(st.In)
                }
                res = append(res, ps)
            }
            return res
        }
        sp.Steps = conv(p.Steps)
        if len(p.ErrorSteps) > 0 { sp.ErrorSteps = conv(p.ErrorSteps) }
        out.Pipelines = append(out.Pipelines, sp)
    }
    return out
}

func toSchemaEdge(s edg.Spec) *sch.PipelineEdgeV1 {
    switch v := s.(type) {
    case edg.FIFO:
        return &sch.PipelineEdgeV1{Kind: v.Kind(), MinCapacity: v.MinCapacity, MaxCapacity: v.MaxCapacity, Backpressure: string(v.Backpressure), Type: v.TypeName}
    case edg.LIFO:
        return &sch.PipelineEdgeV1{Kind: v.Kind(), MinCapacity: v.MinCapacity, MaxCapacity: v.MaxCapacity, Backpressure: string(v.Backpressure), Type: v.TypeName}
    case edg.Pipeline:
        return &sch.PipelineEdgeV1{Kind: v.Kind(), MinCapacity: v.MinCapacity, MaxCapacity: v.MaxCapacity, Backpressure: string(v.Backpressure), Type: v.TypeName, UpstreamName: v.UpstreamName}
    default:
        return &sch.PipelineEdgeV1{Kind: s.Kind()}
    }
}

// parseEdgeSpecFromArgs scans a node's raw arg list and extracts an edge.* spec
// from an `in=` parameter when present.
func parseEdgeSpecFromArgs(args []string) (edg.Spec, bool) {
    for _, a := range args {
        s := strings.TrimSpace(a)
        if !strings.HasPrefix(s, "in=") { continue }
        v := strings.TrimPrefix(s, "in=")
        // Expect one of: edge.FIFO(...), edge.LIFO(...), edge.Pipeline(...)
        if strings.HasPrefix(v, "edge.FIFO(") {
            params := parseKVList(v[len("edge.FIFO(") : len(v)-1])
            var f edg.FIFO
            for k, val := range params {
                switch k {
                case "minCapacity": f.MinCapacity = atoiSafe(val)
                case "maxCapacity": f.MaxCapacity = atoiSafe(val)
                case "backpressure": f.Backpressure = edg.BackpressurePolicy(val)
                case "type": f.TypeName = val
                }
            }
            return f, true
        }
        if strings.HasPrefix(v, "edge.LIFO(") {
            params := parseKVList(v[len("edge.LIFO(") : len(v)-1])
            var l edg.LIFO
            for k, val := range params {
                switch k {
                case "minCapacity": l.MinCapacity = atoiSafe(val)
                case "maxCapacity": l.MaxCapacity = atoiSafe(val)
                case "backpressure": l.Backpressure = edg.BackpressurePolicy(val)
                case "type": l.TypeName = val
                }
            }
            return l, true
        }
        if strings.HasPrefix(v, "edge.Pipeline(") {
            params := parseKVList(v[len("edge.Pipeline(") : len(v)-1])
            var p edg.Pipeline
            for k, val := range params {
                switch k {
                case "name": p.UpstreamName = val
                case "minCapacity": p.MinCapacity = atoiSafe(val)
                case "maxCapacity": p.MaxCapacity = atoiSafe(val)
                case "backpressure": p.Backpressure = edg.BackpressurePolicy(val)
                case "type": p.TypeName = val
                }
            }
            return p, true
        }
    }
    return nil, false
}

// parseKVList parses a simple comma-separated list of `key=value` entries.
// It tolerates identifiers, numbers, quoted strings, and bracketed type tokens
// on the right-hand side. Spaces are not expected (parser strips them).
func parseKVList(s string) map[string]string {
    out := map[string]string{}
    // defensive: balance parentheses if present
    // split on commas at depth 0
    parts := splitTopLevelCommas(s)
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p == "" { continue }
        eq := strings.IndexByte(p, '=')
        if eq < 0 { continue }
        k := strings.TrimSpace(p[:eq])
        v := strings.TrimSpace(p[eq+1:])
        // strip quotes around strings
        if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
            v = v[1:len(v)-1]
        }
        out[k] = v
    }
    return out
}

func splitTopLevelCommas(s string) []string {
    var out []string
    depth := 0
    last := 0
    for i := 0; i < len(s); i++ {
        switch s[i] {
        case '(':
            depth++
        case ')':
            if depth > 0 { depth-- }
        case ',':
            if depth == 0 {
                out = append(out, s[last:i])
                last = i+1
            }
        }
    }
    out = append(out, s[last:])
    return out
}

func atoiSafe(s string) int { n, _ := strconv.Atoi(s); return n }
