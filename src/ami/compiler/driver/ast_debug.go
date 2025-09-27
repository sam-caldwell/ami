package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sort"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

type astUnit struct {
    Schema    string       `json:"schema"`
    Package   string       `json:"package"`
    Unit      string       `json:"unit"`
    Imports   []astImport  `json:"imports,omitempty"`
    Funcs     []astFunc    `json:"funcs,omitempty"`
    Pipelines []astPipe    `json:"pipelines,omitempty"`
}

type astImport struct {
    Path       string `json:"path"`
    Constraint string `json:"constraint,omitempty"`
}

type astTypeParam struct {
    Name       string `json:"name"`
    Constraint string `json:"constraint,omitempty"`
}

type astFunc struct {
    Name       string         `json:"name"`
    TypeParams []astTypeParam `json:"typeParams,omitempty"`
    Params     []string       `json:"params,omitempty"`
    Results    []string       `json:"results,omitempty"`
}

type astPipe struct {
    Name  string       `json:"name"`
    Steps []astPipeStep `json:"steps"`
}

type astPipeStep struct {
    Name string   `json:"name"`
    Args []string `json:"args,omitempty"`
    Attrs []astAttr `json:"attrs,omitempty"`
}

type astAttr struct {
    Name string   `json:"name"`
    Args []string `json:"args,omitempty"`
}

// writeASTDebug writes a per-unit AST summary as ast.v1 JSON.
func writeASTDebug(pkg, unit string, f *ast.File) (string, error) {
    var u astUnit
    u.Schema = "ast.v1"
    u.Package = pkg
    u.Unit = unit
    for _, d := range f.Decls {
        switch n := d.(type) {
        case *ast.ImportDecl:
            u.Imports = append(u.Imports, astImport{Path: n.Path, Constraint: n.Constraint})
        case *ast.FuncDecl:
            var tf []astTypeParam
            for _, tp := range n.TypeParams { tf = append(tf, astTypeParam{Name: tp.Name, Constraint: tp.Constraint}) }
            var ps []string
            for _, p := range n.Params { ps = append(ps, p.Name) }
            var rs []string
            for _, r := range n.Results { rs = append(rs, r.Type) }
            u.Funcs = append(u.Funcs, astFunc{Name: n.Name, TypeParams: tf, Params: ps, Results: rs})
        case *ast.PipelineDecl:
            var steps []astPipeStep
            for _, s := range n.Stmts {
                if st, ok := s.(*ast.StepStmt); ok {
                    var args []string
                    for _, a := range st.Args { args = append(args, a.Text) }
                    // attributes
                    var attrs []astAttr
                    for _, at := range st.Attrs {
                        var aargs []string
                        for _, aa := range at.Args { aargs = append(aargs, aa.Text) }
                        attrs = append(attrs, astAttr{Name: at.Name, Args: aargs})
                    }
                    steps = append(steps, astPipeStep{Name: st.Name, Args: args, Attrs: attrs})
                }
            }
            u.Pipelines = append(u.Pipelines, astPipe{Name: n.Name, Steps: steps})
        }
    }
    // Deterministic ordering
    sort.SliceStable(u.Imports, func(i, j int) bool { return u.Imports[i].Path < u.Imports[j].Path })
    sort.SliceStable(u.Funcs, func(i, j int) bool { return u.Funcs[i].Name < u.Funcs[j].Name })
    sort.SliceStable(u.Pipelines, func(i, j int) bool { return u.Pipelines[i].Name < u.Pipelines[j].Name })
    dir := filepath.Join("build", "debug", "ast", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(u, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, unit+".ast.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}
