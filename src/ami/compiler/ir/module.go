package ir

import (
    "sort"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
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
}

type WorkerIR struct {
    Name       string
    Kind       string // function|factory
    HasContext bool
    HasState   bool
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
                    }
                }
                out = append(out, wi)
            }
            return out
        }
        for _, st := range pd.Steps {
            pir.Steps = append(pir.Steps, StepIR{Node: st.Name, Workers: mkWorkers(st)})
        }
        for _, st := range pd.ErrorSteps {
            pir.ErrorSteps = append(pir.ErrorSteps, StepIR{Node: st.Name, Workers: mkWorkers(st)})
        }
        m.Pipelines = append(m.Pipelines, pir)
    }
}
