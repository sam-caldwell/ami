package driver

import (
    "os"
    "path/filepath"
    "sort"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/sem"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// Compile compiles the provided packages using the given workspace configuration.
// It performs basic memory-safety checks and lowers a minimal imperative subset
// into IR suitable for debug inspection. When opts.Debug is true, it writes
// per-unit IR JSON under build/debug/ir/<package>/<unit>.ir.json.
func Compile(ws workspace.Workspace, pkgs []Package, opts Options) (Artifacts, []diag.Record) {
    var arts Artifacts
    var outDiags []diag.Record
    // process packages in a stable order by name
    sort.SliceStable(pkgs, func(i, j int) bool { return pkgs[i].Name < pkgs[j].Name })
    for _, p := range pkgs {
        if p.Files == nil { continue }
        // collect per-package edges
        var pkgEdges []edgeEntry
        var pkgCollects []collectEntry
        // stable order by file name
        files := append([]*source.File(nil), p.Files.Files...)
        sort.SliceStable(files, func(i, j int) bool { return files[i].Name < files[j].Name })
        for _, f := range files {
            if f == nil { continue }
            // memory safety diagnostics
            outDiags = append(outDiags, sem.AnalyzeMemorySafety(f)...)
            // parse
            pr := parser.New(f)
            af, _ := pr.ParseFile()
            if af == nil { continue }
            // aggregate edges and collect snapshots for package index
            unit := unitName(f.Name)
            pkgEdges = append(pkgEdges, collectEdges(unit, af)...)
            // collect MultiPath snapshots for Collect steps
            for _, d := range af.Decls {
                if pd, ok := d.(*ast.PipelineDecl); ok {
                    for _, s := range pd.Stmts {
                        if st, ok := s.(*ast.StepStmt); ok && st.Name == "Collect" {
                            var mp *edgeMultiPath
                            var merges []mergeAttr
                            var args []string
                            for _, at := range st.Attrs {
                                if at.Name == "edge.MultiPath" || at.Name == "MultiPath" {
                                    for _, a := range at.Args { args = append(args, a.Text) }
                                }
                                if len(at.Name) >= 6 && at.Name[:6] == "merge." {
                                    var margs []string
                                    for _, a := range at.Args { margs = append(margs, a.Text) }
                                    merges = append(merges, mergeAttr{Name: at.Name, Args: margs})
                                }
                            }
                            if len(args) > 0 || len(merges) > 0 {
                                mp = &edgeMultiPath{Args: args, Merge: merges}
                            }
                            if mp != nil {
                                pkgCollects = append(pkgCollects, collectEntry{Unit: unit, Step: st.Name, MultiPath: mp})
                            }
                        }
                    }
                }
            }
            // pipeline and function semantics
            outDiags = append(outDiags, sem.AnalyzePipelineSemantics(af)...)
            outDiags = append(outDiags, sem.AnalyzeFunctions(af)...)
            outDiags = append(outDiags, sem.AnalyzeReturnTypes(af)...)
            // lower functions in this unit into a module
            m := lowerFile(p.Name, af)
            if opts.Debug {
                // write per-unit IR JSON: build/debug/ir/<package>/<unit>.ir.json
                dir := filepath.Join("build", "debug", "ir", p.Name)
                _ = os.MkdirAll(dir, 0o755)
                b, err := ir.EncodeModule(m)
                if err == nil {
                    out := filepath.Join(dir, unit+".ir.json")
                    _ = os.WriteFile(out, b, 0o644)
                    arts.IR = append(arts.IR, out)
                }
                // pipelines debug snapshot
                if _, err := writePipelinesDebug(p.Name, unit, af); err == nil {
                    // intentionally ignore errors in debug artifacts in this scaffold
                }
                // event metadata debug
                if _, err := writeEventMetaDebug(p.Name, unit); err == nil {
                }
                // assembly debug listing
                if _, err := writeAsmDebug(p.Name, unit, m); err == nil {
                }
            }
        }
        if opts.Debug && (len(pkgEdges) > 0 || len(pkgCollects) > 0) {
            _, _ = writeEdgesIndex(p.Name, pkgEdges, pkgCollects)
        }
    }
    return arts, outDiags
}

// unitName is defined in unitname.go
