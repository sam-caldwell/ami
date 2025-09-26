package ir

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    edg "github.com/sam-caldwell/ami/src/ami/compiler/edge"
)

// PipelineIR captures lowered worker/factory references for deeper checks.
type PipelineIR struct {
    Name       string
    Steps      []StepIR
    ErrorSteps []StepIR
}

type StepIR struct {
    Node    string
    Workers []WorkerIR
    In      edg.Spec
}

type WorkerIR struct {
    Name       string
    Kind       string // function|factory
    HasContext bool
    HasState   bool
    Input      string // Event<T> -> T
    OutputKind string // Event|Events|Error
    Output     string // Event<U>/Events<U>/Error<E> -> U/E
}

// LowerPipelines derives PipelineIR using parsed worker refs and function decl signatures.
func (m *Module) LowerPipelines(f *astpkg.File) {
    // build func map
    funs := map[string]astpkg.FuncDecl{}
    for _, d := range f.Decls {
        if fd, ok := d.(astpkg.FuncDecl); ok {
            funs[fd.Name] = fd
        }
    }
    for _, d := range f.Decls {
        pd, ok := d.(astpkg.PipelineDecl)
        if !ok {
            continue
        }
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
                        wi.HasState = (p3.Name == "State")
                        p2 := fd.Params[1].Type
                        if p2.Name == "Event" && len(p2.Args) == 1 {
                            wi.Input = typeRefToString(p2.Args[0])
                        }
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
            if spec, ok := parseEdgeSpecFromArgs(st.Args); ok {
                step.In = spec
            }
            pir.Steps = append(pir.Steps, step)
        }
        for _, st := range pd.ErrorSteps {
            step := StepIR{Node: st.Name, Workers: mkWorkers(st)}
            if spec, ok := parseEdgeSpecFromArgs(st.Args); ok {
                step.In = spec
            }
            pir.ErrorSteps = append(pir.ErrorSteps, step)
        }
        m.Pipelines = append(m.Pipelines, pir)
    }
}
