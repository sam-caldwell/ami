package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// writeSourcesDebug writes a per-unit sources.v1 JSON containing the imports with constraints.
func writeSourcesDebug(pkg, unit string, f *ast.File) (string, error) {
    var list []importDetail
    var prag []pragmaDetail
    for _, d := range f.Decls {
        if im, ok := d.(*ast.ImportDecl); ok { list = append(list, importDetail{Path: im.Path, Constraint: im.Constraint}) }
    }
    for _, pr := range f.Pragmas {
        prag = append(prag, pragmaDetail{Line: pr.Pos.Line, Domain: pr.Domain, Key: pr.Key, Value: pr.Value, Params: pr.Params})
    }
    obj := sourcesUnit{Schema: "sources.v1", Package: pkg, Unit: unit, ImportsDetailed: list, Pragmas: prag}
    dir := filepath.Join("build", "debug", "source", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(obj, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, unit+".sources.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}

