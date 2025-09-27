package driver

import (
    "os"
    "path/filepath"
    "sort"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
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
    var manifestPkgs []bmPackage
    // process packages in a stable order by name
    sort.SliceStable(pkgs, func(i, j int) bool { return pkgs[i].Name < pkgs[j].Name })
    for _, p := range pkgs {
        if p.Files == nil { continue }
        // PHASE 0: sort files deterministically
        files := append([]*source.File(nil), p.Files.Files...)
        sort.SliceStable(files, func(i, j int) bool { return files[i].Name < files[j].Name })
        // memory safety (per file)
        for _, f := range files { if f != nil { outDiags = append(outDiags, sem.AnalyzeMemorySafety(f)...)} }
        // PHASE 1: parse all files and collect ASTs
        type unit struct{ file *source.File; ast *ast.File; unit string }
        var units []unit
        for _, f := range files {
            if f == nil { continue }
            pr := parser.New(f)
            af, _ := pr.ParseFile()
            if af == nil { continue }
            units = append(units, unit{file: f, ast: af, unit: unitName(f.Name)})
        }
        // collect signatures across package
        paramSigs := map[string][]string{}
        paramNames := map[string][]string{}
        resultSigs := map[string][]string{}
        for _, u := range units {
            for _, d := range u.ast.Decls {
                if fn, ok := d.(*ast.FuncDecl); ok && fn.Name != "" {
                    var ps []string
                    var pnames []string
                    var rs []string
                    for _, p := range fn.Params { ps = append(ps, p.Type); pnames = append(pnames, p.Name) }
                    for _, r := range fn.Results { rs = append(rs, r.Type) }
                    paramSigs[fn.Name] = ps
                    paramNames[fn.Name] = pnames
                    resultSigs[fn.Name] = rs
                }
            }
        }
        // PHASE 2: per-unit analyses, edges collection, lowering and debug
        var pkgEdges []edgeEntry
        var pkgCollects []collectEntry
        var bmPkgs []bmPackage
        for _, u := range units {
            af := u.ast
            unit := u.unit
            // manifest package entry lookup/create
            var bmp *bmPackage
            for i := range bmPkgs { if bmPkgs[i].Name == p.Name { bmp = &bmPkgs[i]; break } }
            if bmp == nil { bmPkgs = append(bmPkgs, bmPackage{Name: p.Name}); bmp = &bmPkgs[len(bmPkgs)-1] }
            bmu := bmUnit{Unit: unit}
            // aggregate edges and collect snapshots for package index
            pkgEdges = append(pkgEdges, collectEdges(unit, af)...)
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
                            if len(args) > 0 || len(merges) > 0 { mp = &edgeMultiPath{Args: args, Merge: merges} }
                            if mp != nil { pkgCollects = append(pkgCollects, collectEntry{Unit: unit, Step: st.Name, MultiPath: mp}) }
                        }
                    }
                }
            }
            // analyzers
            attachFile := func(di []diag.Record) {
                for i := range di {
                    if di[i].File == "" {
                        di[i].File = u.file.Name
                    }
                }
                outDiags = append(outDiags, di...)
            }
            attachFile(sem.AnalyzePipelineSemantics(af))
            attachFile(sem.AnalyzeFunctions(af))
            attachFile(sem.AnalyzeMultiPath(af))
            attachFile(sem.AnalyzeEventTypeFlow(af))
            attachFile(sem.AnalyzeReturnTypes(af))
            attachFile(sem.AnalyzeReturnTypesWithSigs(af, resultSigs))
            attachFile(sem.AnalyzeCallsWithSigs(af, paramSigs, resultSigs))
            attachFile(sem.AnalyzePackageAndImports(af))
            attachFile(sem.AnalyzeContainerTypes(af))
            // lower
            m := lowerFile(p.Name, af, paramSigs, resultSigs, paramNames)
            if opts.Debug {
                if s, err := writeSourcesDebug(p.Name, unit, af); err == nil { bmu.Sources = s }
                if a, err := writeASTDebug(p.Name, unit, af); err == nil { bmu.AST = a }
                dir := filepath.Join("build", "debug", "ir", p.Name)
                _ = os.MkdirAll(dir, 0o755)
                b, err := ir.EncodeModule(m)
                if err == nil {
                    out := filepath.Join(dir, unit+".ir.json")
                    _ = os.WriteFile(out, b, 0o644)
                    arts.IR = append(arts.IR, out)
                    bmu.IR = out
                }
                if pp, err := writePipelinesDebug(p.Name, unit, af); err == nil { bmu.Pipelines = pp }
                if em, err := writeEventMetaDebug(p.Name, unit); err == nil { bmu.EventMeta = em }
                if as, err := writeAsmDebug(p.Name, unit, af, m); err == nil { bmu.ASM = as }
                bmp.Units = append(bmp.Units, bmu)
            }
        }
        if opts.Debug && (len(pkgEdges) > 0 || len(pkgCollects) > 0) {
            if ei, err := writeEdgesIndex(p.Name, pkgEdges, pkgCollects); err == nil {
                for i := range bmPkgs { if bmPkgs[i].Name == p.Name { bmPkgs[i].EdgesIndex = ei } }
            }
            if ai, err := writeAsmIndex(p.Name, pkgEdges); err == nil {
                for i := range bmPkgs { if bmPkgs[i].Name == p.Name { bmPkgs[i].AsmIndex = ai } }
            }
        }
        if opts.Debug { manifestPkgs = append(manifestPkgs, bmPkgs...) }
    }
    if opts.Debug && len(manifestPkgs) > 0 {
        _, _ = writeBuildManifest(BuildManifest{Schema: "manifest.v1", Packages: manifestPkgs})
    }
    return arts, outDiags
}

// unitName is defined in unitname.go
