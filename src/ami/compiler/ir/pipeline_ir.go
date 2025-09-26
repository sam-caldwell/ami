package ir

import (
    "strings"

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
	Attrs   map[string]string
}

type WorkerIR struct {
	Name       string
	Kind       string // function|factory
	Origin     string // reference|literal
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
				wi := WorkerIR{Name: w.Name, Kind: w.Kind, Origin: "reference"}
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
			// Inline worker literal (function expression)
			if st.InlineWorker != nil {
				fl := *st.InlineWorker
				wi := WorkerIR{Kind: "inline", Origin: "literal"}
				// Inputs/Outputs based on literal signature (tolerate shorter forms)
				if len(fl.Params) > 0 {
					p := fl.Params[0].Type
					if p.Name == "Event" && len(p.Args) == 1 {
						wi.Input = typeRefToString(p.Args[0])
					}
				}
				if len(fl.Result) > 0 {
					r := fl.Result[0]
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
				// Heuristic: if Params/Result were not populated by parser, derive from attribute text
				if wi.Input == "" || wi.Output == "" {
					if src, ok := st.Attrs["worker"]; ok {
						// extract first "Event<...>" occurrence as input and next as output (if present)
						getPayload := func(s string) (string, string) {
							// naive scan
							var first, second string
							for i := 0; i < len(s); i++ {
								if strings.HasPrefix(s[i:], "Event<") {
									// find matching '>'
									j := i + len("Event<")
									depth := 1
									for j < len(s) {
										if s[j] == '<' {
											depth++
										}
										if s[j] == '>' {
											depth--
											if depth == 0 {
												break
											}
										}
										j++
									}
									if j < len(s) {
										payload := s[i+len("Event<") : j]
										if first == "" {
											first = payload
										} else if second == "" {
											second = payload
											break
										}
										i = j
									}
								}
							}
							return first, second
						}
						in, out := getPayload(src)
						if wi.Input == "" {
							wi.Input = in
						}
						if wi.Output == "" {
							if out != "" {
								wi.OutputKind, wi.Output = "Event", out
							}
						}
					}
				}
				out = append(out, wi)
			}
			return out
		}
        for _, st := range pd.Steps {
            step := StepIR{Node: st.Name, Workers: mkWorkers(st), Attrs: st.Attrs}
            if v := strings.TrimSpace(st.Attrs["in"]); v != "" {
                if spec, ok := parseEdgeSpecFromValue(v); ok { step.In = spec }
            } else if spec, ok := parseEdgeSpecFromArgs(st.Args); ok {
                step.In = spec
            }
            pir.Steps = append(pir.Steps, step)
        }
        for _, st := range pd.ErrorSteps {
            step := StepIR{Node: st.Name, Workers: mkWorkers(st), Attrs: st.Attrs}
            if v := strings.TrimSpace(st.Attrs["in"]); v != "" {
                if spec, ok := parseEdgeSpecFromValue(v); ok { step.In = spec }
            } else if spec, ok := parseEdgeSpecFromArgs(st.Args); ok {
                step.In = spec
            }
            pir.ErrorSteps = append(pir.ErrorSteps, step)
        }
		m.Pipelines = append(m.Pipelines, pir)
	}
}
