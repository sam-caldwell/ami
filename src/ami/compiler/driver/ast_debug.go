package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sort"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

type astUnit struct {
    Schema    string       `json:"schema"`
    Package   string       `json:"package"`
    Unit      string       `json:"unit"`
    Pragmas   []astPragma  `json:"pragmas,omitempty"`
    Imports   []astImport  `json:"imports,omitempty"`
    Funcs     []astFunc    `json:"funcs,omitempty"`
    Pipelines []astPipe    `json:"pipelines,omitempty"`
}

type astImport struct {
    Path       string `json:"path"`
    Constraint string `json:"constraint,omitempty"`
    Pos        *dbgPos `json:"pos,omitempty"`
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
    Decorators []astDecorator `json:"decorators,omitempty"`
    Pos        *dbgPos        `json:"pos,omitempty"`
    NamePos    *dbgPos        `json:"namePos,omitempty"`
}

type astPipe struct {
    Name  string       `json:"name"`
    Steps []astPipeStep `json:"steps"`
    Pos   *dbgPos      `json:"pos,omitempty"`
}

type astPipeStep struct {
    Name string   `json:"name"`
    Args []string `json:"args,omitempty"`
    Attrs []astAttr `json:"attrs,omitempty"`
    Pos  *dbgPos  `json:"pos,omitempty"`
}

type astAttr struct {
    Name string   `json:"name"`
    Args []string `json:"args,omitempty"`
}

type astDecorator struct {
    Name string   `json:"name"`
    Args []string `json:"args,omitempty"`
    Pos  *dbgPos  `json:"pos,omitempty"`
}

type astPragma struct {
    Domain string            `json:"domain,omitempty"`
    Key    string            `json:"key,omitempty"`
    Value  string            `json:"value,omitempty"`
    Args   []string          `json:"args,omitempty"`
    Params map[string]string `json:"params,omitempty"`
    Pos    *dbgPos           `json:"pos,omitempty"`
}

type dbgPos struct {
    Line   int `json:"line"`
    Column int `json:"column"`
    Offset int `json:"offset"`
}

func posPtr(p source.Position) *dbgPos {
    if p.Line == 0 && p.Column == 0 && p.Offset == 0 { return nil }
    return &dbgPos{Line: p.Line, Column: p.Column, Offset: p.Offset}
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
            u.Imports = append(u.Imports, astImport{Path: n.Path, Constraint: n.Constraint, Pos: posPtr(n.Pos)})
        case *ast.FuncDecl:
            var tf []astTypeParam
            for _, tp := range n.TypeParams { tf = append(tf, astTypeParam{Name: tp.Name, Constraint: tp.Constraint}) }
            var ps []string
            for _, p := range n.Params { ps = append(ps, p.Name) }
            var rs []string
            for _, r := range n.Results { rs = append(rs, r.Type) }
            var decos []astDecorator
            for _, d := range n.Decorators {
                var a []string
                for _, e := range d.Args { a = append(a, decoExprText(e)) }
                decos = append(decos, astDecorator{Name: d.Name, Args: a})
            }
            u.Funcs = append(u.Funcs, astFunc{Name: n.Name, TypeParams: tf, Params: ps, Results: rs, Decorators: decos, Pos: posPtr(n.Pos), NamePos: posPtr(n.NamePos)})
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
                    steps = append(steps, astPipeStep{Name: st.Name, Args: args, Attrs: attrs, Pos: posPtr(st.Pos)})
                }
            }
            u.Pipelines = append(u.Pipelines, astPipe{Name: n.Name, Steps: steps, Pos: posPtr(n.Pos)})
        }
    }
    // pragmas
    for _, pr := range f.Pragmas {
        ap := astPragma{Domain: pr.Domain, Key: pr.Key, Value: pr.Value, Pos: posPtr(pr.Pos)}
        if len(pr.Args) > 0 { ap.Args = append(ap.Args, pr.Args...) }
        if len(pr.Params) > 0 { ap.Params = pr.Params }
        u.Pragmas = append(u.Pragmas, ap)
    }
    // Deterministic ordering
    sort.SliceStable(u.Imports, func(i, j int) bool { return u.Imports[i].Path < u.Imports[j].Path })
    sort.SliceStable(u.Pragmas, func(i, j int) bool {
        if u.Pragmas[i].Pos == nil || u.Pragmas[j].Pos == nil { return i < j }
        if u.Pragmas[i].Pos.Line == u.Pragmas[j].Pos.Line { return u.Pragmas[i].Pos.Column < u.Pragmas[j].Pos.Column }
        return u.Pragmas[i].Pos.Line < u.Pragmas[j].Pos.Line
    })
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

// decoExprText is a tiny helper to stringify AST expressions for decorator args.
// Keep this aligned with parser's exprText for consistency in debug outputs.
func decoExprText(e ast.Expr) string {
    switch v := e.(type) {
    case *ast.IdentExpr:
        return v.Name
    case *ast.StringLit:
        return v.Value
    case *ast.NumberLit:
        return v.Text
    case *ast.SelectorExpr:
        left := decoExprText(v.X)
        if left == "" { left = "?" }
        return left + "." + v.Sel
    case *ast.CallExpr:
        if len(v.Args) > 0 { return v.Name + "(…)" }
        return v.Name + "()"
    case *ast.SliceLit:
        return "slice"
    case *ast.SetLit:
        return "set"
    case *ast.MapLit:
        return "map"
    default:
        return ""
    }
}
