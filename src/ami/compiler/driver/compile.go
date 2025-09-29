package driver

import (
    "os"
    "path/filepath"
    "sort"
    "time"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/codegen"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/sem"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/schemas/diag"
    "encoding/json"
)

// small helper for human one-liners without extra deps
func joinCSV(ss []string) string { if len(ss)==0 { return "" }; out := ss[0]; for i:=1; i<len(ss); i++ { out += "," + ss[i] }; return out }

// Compile compiles the provided packages using the given workspace configuration.
// It performs basic memory-safety checks and lowers a minimal imperative subset
// into IR suitable for debug inspection. When opts.Debug is true, it writes
// per-unit IR JSON under build/debug/ir/<package>/<unit>.ir.json.
func Compile(ws workspace.Workspace, pkgs []Package, opts Options) (Artifacts, []diag.Record) {
    var arts Artifacts
    var outDiags []diag.Record
    var manifestPkgs []bmPackage
    if opts.Log != nil {
        opts.Log("start", map[string]any{"packages": len(pkgs), "debug": opts.Debug})
    }
    // process packages in a stable order by name
    sort.SliceStable(pkgs, func(i, j int) bool { return pkgs[i].Name < pkgs[j].Name })
    // collect resolved sources for debug summary
    var resolved []resolvedUnit
    // Apply workspace-driven decorator disables (scaffold wiring for analyzers)
    if len(ws.Toolchain.Linter.DecoratorsDisabled) > 0 {
        sem.SetDisabledDecorators(ws.Toolchain.Linter.DecoratorsDisabled...)
    } else {
        sem.SetDisabledDecorators()
    }
    for _, p := range pkgs {
        if opts.Log != nil { opts.Log("pkg.start", map[string]any{"pkg": p.Name}) }
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
            af, perrs := pr.ParseFileCollect()
            // convert syntax errors to diagnostics
            for _, e := range perrs {
                // default position
                pos := source.Position{}
                if se, ok := e.(parser.SyntaxError); ok { pos = se.Position() }
                outDiags = append(outDiags, diag.Record{
                    Timestamp: time.Now().UTC(),
                    Level:     diag.Error,
                    Code:      "E_PARSE_SYNTAX",
                    Message:   e.Error(),
                    File:      f.Name,
                    Pos:       &diag.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
                })
            }
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
        // Package-level pipeline egress type map for edge.Pipeline resolution across units
        egressTypesPkg := map[string]string{}
        for _, u := range units {
            if u.ast == nil { continue }
            for _, d := range u.ast.Decls {
                if pd, ok := d.(*ast.PipelineDecl); ok {
                    t := ""
                    for _, s := range pd.Stmts {
                        if st, ok := s.(*ast.StepStmt); ok {
                            if strings.ToLower(st.Name) == "egress" {
                                for _, at := range st.Attrs {
                                    if at.Name == "type" || at.Name == "Type" {
                                        if len(at.Args) > 0 { t = at.Args[0].Text }
                                    }
                                }
                            }
                        }
                    }
                    egressTypesPkg[pd.Name] = t
                }
            }
        }
        var pkgEdges []edgeEntry
        var pkgCollects []collectEntry
        var bmPkgs []bmPackage
        var irUnits []irIndexUnit
        var typesUnits []irTypesIndexUnit
        var symbolsUnits []irSymbolsIndexUnit
        for _, u := range units {
            if opts.Log != nil { opts.Log("unit.start", map[string]any{"pkg": p.Name, "unit": u.unit}) }
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
            attachFile(sem.AnalyzeDecorators(af))
            attachFile(sem.AnalyzeMultiPath(af))
            attachFile(sem.AnalyzeEnums(af))
            attachFile(sem.AnalyzeConcurrency(af))
            attachFile(sem.AnalyzeEdgesInContext(af, egressTypesPkg))
            attachFile(sem.AnalyzeEventTypeFlow(af))
            attachFile(sem.AnalyzeContainerTypes(af))
            attachFile(sem.AnalyzeTypeInference(af))
            attachFile(sem.AnalyzeAmbiguity(af))
            attachFile(sem.AnalyzeNameResolution(af))
            attachFile(sem.AnalyzeWorkers(af))
            attachFile(sem.AnalyzeReturnTypes(af))
            // Include return-type inference for functions without annotations (M8 scope)
            attachFile(sem.AnalyzeReturnInference(af))
            attachFile(sem.AnalyzeReturnTypesWithSigs(af, resultSigs))
            attachFile(sem.AnalyzeRAII(af))
            attachFile(sem.AnalyzeCallsWithSigs(af, paramSigs, resultSigs))
            attachFile(sem.AnalyzePackageAndImports(af))
            // IR/codegen-stage capability check (complements semantics layer)
            attachFile(analyzeCapabilityIR(af))
            attachFile(sem.AnalyzeContainerTypes(af))
            // lower
            m := lowerFile(p.Name, af, paramSigs, resultSigs, paramNames)
            // Optimizer M18 (DCE): remove unreferenced functions per file
            // Only apply when a 'main' root exists; otherwise keep all functions for tooling/tests.
            hasMain := false
            for _, d := range af.Decls { if fn, ok := d.(*ast.FuncDecl); ok && fn.Name == "main" { hasMain = true; break } }
            if hasMain {
                reach := sem.ReachableFunctions(af)
                if len(reach) > 0 {
                    var kept []ir.Function
                    for _, fn := range m.Functions { if reach[fn.Name] { kept = append(kept, fn) } }
                    if len(kept) > 0 { m.Functions = kept }
                }
            }
            if opts.Debug {
                if s, err := writeSourcesDebug(p.Name, unit, af); err == nil { bmu.Sources = s }
                // accumulate resolved sources payload for top-level summary
                var imports []string
                for _, d := range af.Decls { if im, ok := d.(*ast.ImportDecl); ok { imports = append(imports, im.Path) } }
                resolved = append(resolved, resolvedUnit{Package: p.Name, File: u.file.Name, Imports: imports, Source: u.file.Content})
                if a, err := writeASTDebug(p.Name, unit, af); err == nil { bmu.AST = a }
                dir := filepath.Join("build", "debug", "ir", p.Name)
                _ = os.MkdirAll(dir, 0o755)
                b, err := ir.EncodeModule(m)
                if err == nil {
                    out := filepath.Join(dir, unit+".ir.json")
                    _ = os.WriteFile(out, b, 0o644)
                    arts.IR = append(arts.IR, out)
                    bmu.IR = out
                    if opts.Log != nil { opts.Log("unit.ir.write", map[string]any{"pkg": p.Name, "unit": unit, "path": out}) }
                }
                if pp, err := writePipelinesDebug(p.Name, unit, af); err == nil {
                    bmu.Pipelines = pp
                    // In verbose mode, emit human-friendly connectivity summaries.
                    if opts.Log != nil {
                        // Parse the pipelines JSON to extract connectivity for logging.
                        type pipeConn struct{
                            HasEdges bool `json:"hasEdges"`
                            IngressToEgress bool `json:"ingressToEgress"`
                            Disconnected []string `json:"disconnected"`
                            UnreachableFromIngress []string `json:"unreachableFromIngress"`
                            CannotReachEgress []string `json:"cannotReachEgress"`
                        }
                        type pipeEntry struct{
                            Name string `json:"name"`
                            Edges []any `json:"edges"`
                            Conn *pipeConn `json:"connectivity"`
                        }
                        var obj struct{ Pipelines []pipeEntry `json:"pipelines"` }
                        if b, rerr := os.ReadFile(pp); rerr == nil {
                            if jerr := json.Unmarshal(b, &obj); jerr == nil {
                                for _, pe := range obj.Pipelines {
                                    if pe.Conn != nil {
                                        fields := map[string]any{
                                            "pkg": p.Name, "unit": unit, "pipeline": pe.Name,
                                            "edges": len(pe.Edges), "ingressToEgress": pe.Conn.IngressToEgress,
                                        }
                                        if len(pe.Conn.Disconnected) > 0 { fields["disconnected"] = pe.Conn.Disconnected }
                                        if len(pe.Conn.UnreachableFromIngress) > 0 { fields["unreachableFromIngress"] = pe.Conn.UnreachableFromIngress }
                                        if len(pe.Conn.CannotReachEgress) > 0 { fields["cannotReachEgress"] = pe.Conn.CannotReachEgress }
                                        opts.Log("unit.pipelines.connectivity", fields)
                                        // Human one-liner summary for activity.log readers
                                        // Example: "pipeline P: edges=3, ingress→egress=false, unreachable=[B], nonterm=[A]"
                                        line := "pipeline " + pe.Name + ": edges=" + itoa(len(pe.Edges)) + ", ingress→egress="
                                        if pe.Conn.IngressToEgress { line += "true" } else { line += "false" }
                                        if len(pe.Conn.UnreachableFromIngress) > 0 {
                                            line += ", unreachable=[" + joinCSV(pe.Conn.UnreachableFromIngress) + "]"
                                        }
                                        if len(pe.Conn.CannotReachEgress) > 0 {
                                            line += ", nonterm=[" + joinCSV(pe.Conn.CannotReachEgress) + "]"
                                        }
                                        opts.Log("unit.pipelines.connectivity.human", map[string]any{"pkg": p.Name, "unit": unit, "pipeline": pe.Name, "text": line})
                                    }
                                }
                            }
                        }
                    }
                }
                if ct, err := writeContractsDebug(p.Name, unit, af); err == nil { bmu.Contracts = ct }
                if ssa, err := writeSSADebug(p.Name, unit, m); err == nil { _ = ssa }
                if em, err := writeEventMetaDebug(p.Name, unit); err == nil { bmu.EventMeta = em }
                if as, err := writeAsmDebug(p.Name, unit, af, m); err == nil { bmu.ASM = as }
                if _, err := writeExportsDebug(p.Name, unit, m); err == nil { /* ok: optional debug */ }
                // accumulate IR index info for this unit
                var fnames []string
                for _, fn := range m.Functions { fnames = append(fnames, fn.Name) }
                irUnits = append(irUnits, irIndexUnit{Unit: unit, Functions: fnames})
                typesUnits = append(typesUnits, irTypesIndexUnit{Unit: unit, Types: collectTypes(m)})
                symbolsUnits = append(symbolsUnits, irSymbolsIndexUnit{Unit: unit, Exports: collectExports(m), Externs: collectExterns(m)})
            // IR textual emission (backend; debug only)
            be := codegen.DefaultBackend()
            if llvmText, err := be.EmitModule(m); err == nil {
                ldir := filepath.Join("build", "debug", "llvm", p.Name)
                _ = os.MkdirAll(ldir, 0o755)
                lpath := filepath.Join(ldir, unit+".ll")
                _ = os.WriteFile(lpath, []byte(llvmText), 0o644)
                bmu.LLVM = lpath
                if opts.Log != nil { opts.Log("unit.llvm.write", map[string]any{"pkg": p.Name, "unit": unit, "path": lpath}) }
                // Attempt to compile to an object even in debug builds so objindex lists real .o when toolchain is present.
                if !opts.EmitLLVMOnly {
                    if clang, err := be.FindToolchain(); err != nil {
                        outDiags = append(outDiags, diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_TOOLCHAIN_MISSING", Message: "clang not found in PATH", File: "clang"})
                    } else {
                        objDir := filepath.Join("build", "obj", p.Name)
                        _ = os.MkdirAll(objDir, 0o755)
                        oPath := filepath.Join(objDir, unit+".o")
                        if err := be.CompileLLToObject(clang, lpath, oPath, ""); err != nil {
                            if opts.DebugStrict {
                                d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_OBJ_COMPILE_FAIL", Message: "failed to compile LLVM to object", File: lpath}
                                if te, ok := err.(interface{ Error() string; String() string }); ok {
                                    _ = te // placeholder to avoid unused; below we'll use ToolError detection
                                }
                                // Best effort: if backend exposes stderr via known type
                                type toolErr interface{ Stderr() string }
                                if te, ok := any(err).(toolErr); ok {
                                    if d.Data == nil { d.Data = map[string]any{} }
                                    d.Data["stderr"] = te.Stderr()
                                }
                                outDiags = append(outDiags, d)
                            } else if opts.Log != nil {
                                opts.Log("unit.obj.debug.compile.fail", map[string]any{"pkg": p.Name, "unit": unit, "error": err.Error()})
                            }
                        } else if opts.Log != nil {
                            opts.Log("unit.obj.write", map[string]any{"pkg": p.Name, "unit": unit, "path": oPath})
                        }
                    }
                }
            } else {
                // surface emission errors in debug as diagnostics too
                outDiags = append(outDiags, diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_LLVM_EMIT", Message: err.Error(), File: unit + ".ll"})
            }
                // Optional: RAII trace debug for this unit
                if rpath, err := writeIRRAIIDebug(p.Name, unit, af); err == nil { bmu.RAII = rpath }
                bmp.Units = append(bmp.Units, bmu)
            }
            // emit object stub and per-unit asm under build/obj in all modes
            if !opts.EmitLLVMOnly {
                _, _ = writeObjectStub(p.Name, unit, m)
            }
            _, _ = writeAsmObject(p.Name, unit, m)
            // Attempt non-debug IR -> .o compilation (guarded)
            // Emit .ll and attempt to compile to .o. When env matrix is present, emit strictly per-env.
            if !opts.Debug {
                be := codegen.DefaultBackend()
                if clang, err := be.FindToolchain(); err != nil {
                    outDiags = append(outDiags, diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_TOOLCHAIN_MISSING", Message: "clang not found in PATH", File: "clang"})
                } else {
                    // Emit per-env objects under build/<env>/obj/<pkg>/unit.o
                    for _, env := range ws.Toolchain.Compiler.Env {
                        triple := be.TripleForEnv(env)
                        if llvmText, err := be.EmitModuleForTarget(m, triple); err == nil {
                            envObjDir := filepath.Join("build", env, "obj", p.Name)
                            _ = os.MkdirAll(envObjDir, 0o755)
                            llPathEnv := filepath.Join(envObjDir, unit+".ll")
                            _ = os.WriteFile(llPathEnv, []byte(llvmText), 0o644)
                            if !opts.EmitLLVMOnly {
                                oPathEnv := filepath.Join(envObjDir, unit+".o")
                                if err := be.CompileLLToObject(clang, llPathEnv, oPathEnv, triple); err == nil {
                                    if opts.Log != nil { opts.Log("unit.obj.env.write", map[string]any{"pkg": p.Name, "unit": unit, "env": env, "path": oPathEnv}) }
                                }
                            }
                        } else {
                            outDiags = append(outDiags, diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_LLVM_EMIT", Message: err.Error(), File: unit + ".ll"})
                        }
                    }
                    // If no env matrix provided, emit default object under build/obj/<pkg>
                    if len(ws.Toolchain.Compiler.Env) == 0 {
                        defTriple := ""
                        if llvmText, err := be.EmitModuleForTarget(m, defTriple); err == nil {
                            objDir := filepath.Join("build", "obj", p.Name)
                            _ = os.MkdirAll(objDir, 0o755)
                            llPath := filepath.Join(objDir, unit+".ll")
                            _ = os.WriteFile(llPath, []byte(llvmText), 0o644)
                            if opts.Log != nil { opts.Log("unit.ll.emit", map[string]any{"pkg": p.Name, "unit": unit, "path": llPath}) }
                            if !opts.EmitLLVMOnly {
                                oPath := filepath.Join(objDir, unit+".o")
                                if err := be.CompileLLToObject(clang, llPath, oPath, defTriple); err == nil {
                                    if opts.Log != nil { opts.Log("unit.obj.write", map[string]any{"pkg": p.Name, "unit": unit, "path": oPath}) }
                                }
                            }
                        } else {
                            outDiags = append(outDiags, diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_LLVM_EMIT", Message: err.Error(), File: unit + ".ll"})
                        }
                    } else if opts.EmitLLVMOnly {
                        // Honor test contract: when EmitLLVMOnly and envs present, still emit a default .ll under build/obj for inspection.
                        if llvmText, err := be.EmitModuleForTarget(m, ""); err == nil {
                            objDir := filepath.Join("build", "obj", p.Name)
                            _ = os.MkdirAll(objDir, 0o755)
                            llPath := filepath.Join(objDir, unit+".ll")
                            _ = os.WriteFile(llPath, []byte(llvmText), 0o644)
                            if opts.Log != nil { opts.Log("unit.ll.emit", map[string]any{"pkg": p.Name, "unit": unit, "path": llPath}) }
                        }
                    }
                }
            }
            if opts.Log != nil { opts.Log("unit.end", map[string]any{"pkg": p.Name, "unit": unit}) }
        }
        if opts.Debug && (len(pkgEdges) > 0 || len(pkgCollects) > 0 || len(irUnits) > 0) {
            if ei, err := writeEdgesIndex(p.Name, pkgEdges, pkgCollects); err == nil {
                for i := range bmPkgs { if bmPkgs[i].Name == p.Name { bmPkgs[i].EdgesIndex = ei } }
            }
            if ai, err := writeAsmIndex(p.Name, pkgEdges); err == nil {
                for i := range bmPkgs { if bmPkgs[i].Name == p.Name { bmPkgs[i].AsmIndex = ai } }
            }
            if ii, err := writeIRIndex(p.Name, irUnits); err == nil {
                for i := range bmPkgs { if bmPkgs[i].Name == p.Name { bmPkgs[i].IRIndex = ii } }
            }
            if ti, err := writeIRTypesIndex(p.Name, typesUnits); err == nil {
                for i := range bmPkgs { if bmPkgs[i].Name == p.Name { bmPkgs[i].IRTypesIndex = ti } }
            }
            if si, err := writeIRSymbolsIndex(p.Name, symbolsUnits); err == nil {
                for i := range bmPkgs { if bmPkgs[i].Name == p.Name { bmPkgs[i].IRSymbolsIndex = si } }
            }
        }
        // Build object index for package under build/obj/<pkg> (always)
        objDir := filepath.Join("build", "obj", p.Name)
        if idx, err := codegen.BuildObjIndex(p.Name, objDir); err == nil { _ = codegen.WriteObjIndex(idx) }
        // Build per-env object indexes under build/<env>/obj/<pkg>/ when present
        for _, env := range ws.Toolchain.Compiler.Env {
            envObjDir := filepath.Join("build", env, "obj", p.Name)
            if st, err := os.Stat(envObjDir); err == nil && st.IsDir() {
                if idx, err := codegen.BuildObjIndex(p.Name, envObjDir); err == nil {
                    // Write index into the env-specific folder
                    b, _ := json.MarshalIndent(idx, "", "  ")
                    _ = os.WriteFile(filepath.Join(envObjDir, "index.json"), b, 0o644)
                }
            }
        }
        if opts.Debug { manifestPkgs = append(manifestPkgs, bmPkgs...) }
        if opts.Log != nil { opts.Log("pkg.end", map[string]any{"pkg": p.Name}) }
    }
    if opts.Debug && len(manifestPkgs) > 0 {
        _, _ = writeBuildManifest(BuildManifest{Schema: "manifest.v1", Packages: manifestPkgs})
    }
    if opts.Debug && len(resolved) > 0 {
        _, _ = writeResolvedSourcesDebug(resolved)
    }
    if opts.Log != nil { opts.Log("end", map[string]any{"packages": len(pkgs), "debug": opts.Debug}) }
    return arts, outDiags
}

// unitName is defined in unitname.go
