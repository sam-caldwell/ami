package main

import (
    "bufio"
    "bytes"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/sem"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// lintStageB invokes parser/semantics-backed rules when enabled by toggles t.
// Currently implements MemorySafety diagnostics over all .ami files reachable
// from the main package root and its recursive local imports.
func lintStageB(dir string, ws *workspace.Workspace, t RuleToggles) []diag.Record {
    var out []diag.Record
    if ws == nil { return out }

    // Determine package roots to scan: child-first local imports then main root.
    roots := []string{}
    if p := ws.FindPackage("main"); p != nil && p.Root != "" {
        roots = append(collectLocalImportRoots(ws, p), p.Root)
    }
    if len(roots) == 0 { return out }

    // Gather pragma disables across all roots so we can filter emitted records.
    disables := map[string]map[string]bool{}
    for _, r := range roots {
        m := scanPragmas(dir, r)
        for file, rules := range m { disables[file] = rules }
    }

    // Walk each root, analyze every .ami file.
    for _, r := range roots {
        root := filepath.Clean(filepath.Join(dir, r))
        // Track duplicates across files within this root
        dupFuncs := map[string]sourceWithPos{}
        _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
            if err != nil || d.IsDir() || filepath.Ext(path) != ".ami" { return nil }
            // Read file content and create a source.File
            b, err := os.ReadFile(path)
            if err != nil { return nil }
            f := &source.File{Name: path, Content: string(b)}
            now := time.Now().UTC()

            // Memory safety without external sem package dependency
            if t.StageB || t.MemorySafety {
                ms := analyzeMemorySafety(f)
                if len(ms) > 0 {
                    for _, d := range ms {
                        if m := disables[path]; m != nil && m[d.Code] { continue }
                        out = append(out, d)
                    }
                }
            }

            // Parser-backed rules
            pf := parser.New(f)
            af, err := pf.ParseFile()
            if err != nil {
                // tolerate parse errors in Stage B: emit a diag and continue
                d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_PARSE_TOLERANT", Message: "parse error in Stage B lint: " + err.Error(), File: path}
                if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                return nil
            }

            // Bridge selected semantics diagnostics into lint when Stage B is enabled.
            if t.StageB {
                // Apply decorator disables from workspace config
                if names := ws.Toolchain.Linter.DecoratorsDisabled; len(names) > 0 {
                    sem.SetDisabledDecorators(names...)
                } else {
                    sem.SetDisabledDecorators()
                }
                // Propagate strictness toggles: CLI override (if set) wins, else workspace config
                if t.HasStrictMDPOverride {
                    sem.SetStrictDedupUnderPartition(t.StrictMDPOverride)
                } else {
                    sem.SetStrictDedupUnderPartition(ws.Toolchain.Linter.StrictMergeDedupPartition)
                }
                semDiags := append(sem.AnalyzePipelineSemantics(af), sem.AnalyzeErrorSemantics(af)...)
                semDiags = append(semDiags, sem.AnalyzeDecorators(af) ...)
                for _, sd := range semDiags {
                    d := sd
                    if d.File == "" { d.File = path }
                    if m := disables[path]; m != nil && m[d.Code] { continue }
                    out = append(out, d)
                }
            }

            // Unused detection: variables and functions (file-local)
            if t.StageB || t.Unused {
                for _, ud := range sem.AnalyzeUnused(af) {
                    d := ud
                    if d.File == "" { d.File = path }
                    if m := disables[path]; m != nil && m[d.Code] { continue }
                    out = append(out, d)
                }
            }

            // Collect used identifiers (first segment of selector/call names) for unused import checks.
            used := collectUsedIdents(af)

            // Unused imports: only for identifier-form imports (not strings)
            if t.StageB || t.Unused || t.ImportExist {
                seenAlias := map[string]sourceWithPos{}
                for _, dcl := range af.Decls {
                    im, ok := dcl.(*ast.ImportDecl); if !ok { continue }
                    // Duplicate alias detection: explicit alias used more than once
                    if im.Alias != "" {
                        if prior, seen := seenAlias[im.Alias]; seen {
                            d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_DUP_IMPORT_ALIAS", Message: "duplicate import alias: " + im.Alias, File: path, Pos: &diag.Position{Line: im.AliasPos.Line, Column: im.AliasPos.Column, Offset: im.AliasPos.Offset}, Data: map[string]any{"prevLine": prior.line, "prevColumn": prior.col}}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        } else {
                            seenAlias[im.Alias] = sourceWithPos{line: im.AliasPos.Line, col: im.AliasPos.Column}
                        }
                    }
                    // Identifier-form imports only (ignore string imports)
                    if im.Path != "" && !strings.Contains(im.Path, "/") {
                        name := im.Path
                        if !used[name] {
                            d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_UNUSED_IMPORT", Message: "import not used: " + name, File: path, Pos: &diag.Position{Line: im.PathPos.Line, Column: im.PathPos.Column, Offset: im.PathPos.Offset}}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        }
                    }
                }
            }

            // Unmarked mutating assignment: any assignment statement (since '* name = ...' is not parsed as AssignStmt)
            if t.StageB || t.MemorySafety {
                walkStmts(af, func(s ast.Stmt) {
                    if as, ok := s.(*ast.AssignStmt); ok {
                        d := diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MUT_ASSIGN_UNMARKED", Message: "assignment should use mutating marker: '* name = expr'", File: path, Pos: &diag.Position{Line: as.NamePos.Line, Column: as.NamePos.Column, Offset: as.NamePos.Offset}}
                        if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                    }
                })
            }

            // Duplicate function declarations across files in this root
            if t.StageB {
                for _, dcl := range af.Decls {
                    if fn, ok := dcl.(*ast.FuncDecl); ok && fn != nil {
                        name := fn.Name
                        if prior, seen := dupFuncs[name]; seen {
                            d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_DUP_FUNC_DECL", Message: "duplicate function declaration: " + name, File: path, Pos: &diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset}, Data: map[string]any{"prevLine": prior.line, "prevColumn": prior.col}}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        } else {
                            dupFuncs[name] = sourceWithPos{line: fn.NamePos.Line, col: fn.NamePos.Column}
                        }
                    }
                }
            }

            // Pipeline position hints: ingress should be first; egress should be last
            if t.StageB {
                for _, dcl := range af.Decls {
                    pd, ok := dcl.(*ast.PipelineDecl); if !ok || pd == nil { continue }
                    var stmts []ast.Stmt
                    if len(pd.Stmts) > 0 { stmts = pd.Stmts } else if pd.Body != nil { stmts = pd.Body.Stmts }
                    if len(stmts) == 0 { continue }
                    firstIdx := 0
                    lastIdx := len(stmts) - 1
                    ingressIdx := -1
                    egressIdx := -1
                    var ingressPos, egressPos diag.Position
                    for i, s := range stmts {
                        if st, ok := s.(*ast.StepStmt); ok {
                            lname := strings.ToLower(st.Name)
                            if lname == "ingress" && ingressIdx < 0 { ingressIdx = i; ingressPos = diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset} }
                            if lname == "egress" { egressIdx = i; egressPos = diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset} }
                        }
                    }
                    if ingressIdx >= 0 && ingressIdx != firstIdx {
                        d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_PIPELINE_INGRESS_POS", Message: "'ingress' should be the first step in a pipeline", File: path, Pos: &ingressPos}
                        if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                    }
                    if egressIdx >= 0 && egressIdx != lastIdx {
                        d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_PIPELINE_EGRESS_POS", Message: "'egress' should be the last step in a pipeline", File: path, Pos: &egressPos}
                        if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                    }
                }
            }

            // Collections hints
            if t.StageB {
                walkExprs(af, func(e ast.Expr) {
                    switch n := e.(type) {
                    case *ast.SliceLit:
                        if len(n.Elems) == 1 {
                            d := diag.Record{Timestamp: now, Level: diag.Info, Code: "W_SLICE_ARITY_HINT", Message: "slice literal has a single element; verify intended arity", File: path, Pos: &diag.Position{Line: n.Pos.Line, Column: n.Pos.Column, Offset: n.Pos.Offset}, Data: map[string]any{"count": 1}}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        }
                    case *ast.SetLit:
                        if len(n.Elems) == 0 {
                            d := diag.Record{Timestamp: now, Level: diag.Info, Code: "W_SET_EMPTY_HINT", Message: "set literal is empty", File: path, Pos: &diag.Position{Line: n.Pos.Line, Column: n.Pos.Column, Offset: n.Pos.Offset}}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        }
                    case *ast.MapLit:
                        if len(n.Elems) == 0 {
                            d := diag.Record{Timestamp: now, Level: diag.Info, Code: "W_MAP_EMPTY_HINT", Message: "map literal is empty", File: path, Pos: &diag.Position{Line: n.Pos.Line, Column: n.Pos.Column, Offset: n.Pos.Offset}}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        }
                    }
                })
            }

            // Pipeline buffer/backpressure smells, capability (I/O) checks, and reachability
            if t.StageB {
                // Detect Buffer(...) policy smell and alias usage via raw source scan (best-effort)
                lowsrc := strings.ToLower(string(b))
                if idx := strings.Index(lowsrc, "buffer("); idx >= 0 {
                    // dropNewest spelled-out policy
                    if j := strings.Index(lowsrc[idx:], "dropnewest"); j >= 0 {
                        pos := idx + j
                        // compute approximate line/column
                        line, col := 1, 1
                        for k := 0; k < pos && k < len(b); k++ { if b[k] == '\n' { line++; col = 1 } else { col++ } }
                        d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_PIPELINE_BUFFER_POLICY_SMELL", Message: "Buffer policy may drop newest items; verify appropriateness", File: path, Pos: &diag.Position{Line: line, Column: col, Offset: pos}}
                        if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                    }
                    // alias 'drop' policy
                    if j := strings.Index(lowsrc[idx:], "drop)"); j >= 0 {
                        pos := idx + j
                        line, col := 1, 1
                        for k := 0; k < pos && k < len(b); k++ { if b[k] == '\n' { line++; col = 1 } else { col++ } }
                        d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_PIPELINE_BUFFER_DROP_ALIAS", Message: "Buffer policy alias 'drop' used; prefer explicit policy", File: path, Pos: &diag.Position{Line: line, Column: col, Offset: pos}}
                        if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                    }
                }
                // Gather declared capabilities, trust level, and concurrency hints from pragmas
                declCaps := map[string]bool{}
                trustLevel := ""
                workers := 0
                schedule := ""
                limits := map[string]int{}
                for _, pr := range af.Pragmas {
                    switch strings.ToLower(pr.Domain) {
                    case "capabilities":
                        if pr.Params != nil {
                            if lst, ok := pr.Params["list"]; ok && lst != "" {
                                for _, p := range strings.Split(lst, ",") {
                                    p = strings.TrimSpace(p)
                                    if p != "" { declCaps[strings.ToLower(p)] = true }
                                }
                            }
                        }
                        for _, a := range pr.Args { declCaps[strings.ToLower(a)] = true }
                    case "trust":
                        if pr.Params != nil {
                            if lv, ok := pr.Params["level"]; ok { trustLevel = strings.ToLower(lv) }
                        }
                    case "concurrency":
                        switch strings.ToLower(pr.Key) {
                        case "workers":
                            // value or params[count]
                            if pr.Value != "" { workers = atoiSafe(pr.Value) } else if pr.Params != nil { if v, ok := pr.Params["count"]; ok { workers = atoiSafe(v) } }
                        case "schedule":
                            schedule = strings.ToLower(pr.Value)
                        case "limits":
                            // collect numeric args/params
                            for k, v := range pr.Params { if n := atoiSafe(v); n > 0 { limits[strings.ToLower(k)] = n } }
                            for _, a := range pr.Args {
                                if eq := strings.IndexByte(a, '='); eq > 0 { k := strings.ToLower(strings.TrimSpace(a[:eq])); v := strings.TrimSpace(a[eq+1:]); if n := atoiSafe(v); n > 0 { limits[k] = n } }
                            }
                        }
                    }
                }
                // Fallback: scan raw source pragmas if parser did not attach them
                if len(declCaps) == 0 || trustLevel == "" {
                    scanner := bufio.NewScanner(bytes.NewReader(b))
                    for scanner.Scan() {
                        ln := strings.TrimSpace(scanner.Text())
                        if !strings.HasPrefix(ln, "#pragma ") { continue }
                        body := strings.TrimSpace(strings.TrimPrefix(ln, "#pragma "))
                        if strings.HasPrefix(body, "capabilities ") {
                            rest := strings.TrimSpace(strings.TrimPrefix(body, "capabilities "))
                            if strings.HasPrefix(rest, "list=") {
                                kv := parseKV(rest)
                                if lst := kv["list"]; lst != "" {
                                    for _, p := range strings.Split(lst, ",") {
                                        p = strings.TrimSpace(p)
                                        if p != "" { declCaps[strings.ToLower(p)] = true }
                                    }
                                }
                            } else {
                                // treat remaining tokens as args
                                for _, tok := range strings.Fields(rest) {
                                    t := strings.Trim(tok, "\"'")
                                    if t != "" { declCaps[strings.ToLower(t)] = true }
                                }
                            }
                        } else if strings.HasPrefix(body, "trust ") {
                            rest := strings.TrimSpace(strings.TrimPrefix(body, "trust "))
                            kv := parseKV(rest)
                            if lv := kv["level"]; lv != "" { trustLevel = strings.ToLower(lv) }
                        } else if strings.HasPrefix(body, "concurrency ") {
                            rest := strings.TrimSpace(strings.TrimPrefix(body, "concurrency "))
                            kv := parseKV(rest)
                            if v := kv["workers"]; v != "" { if n := atoiSafe(v); n > 0 { workers = n } }
                            if v := kv["schedule"]; v != "" { schedule = strings.ToLower(v) }
                            for k, v := range kv { if n := atoiSafe(v); n > 0 { limits[strings.ToLower(k)] = n } }
                        }
                    }
                }
                // Evaluate each pipeline: check capability/position/policy hints
                for _, dcl := range af.Decls {
                    pd, ok := dcl.(*ast.PipelineDecl); if !ok || pd == nil { continue }
                    var stmts []ast.Stmt
                    if len(pd.Stmts) > 0 { stmts = pd.Stmts } else if pd.Body != nil { stmts = pd.Body.Stmts }
                    if len(stmts) == 0 { continue }
                    // Allowed I/O checks: ingress/egress only
                    for _, s := range stmts {
                        st, ok := s.(*ast.StepStmt); if !ok { continue }
                        lname := strings.ToLower(st.Name)
                        pos := "middle"
                        if st == stmts[0] { pos = "ingress" } else if st == stmts[len(stmts)-1] { pos = "egress" }
                        if strings.HasPrefix(lname, "io.") {
                            if pos == "ingress" && !ioAllowedIngress(lname) {
                                d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_IO_INGRESS_UNSAFE", Message: "unsafe ingress I/O operation: " + st.Name, File: path, Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}}
                                if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                            }
                            if pos == "egress" && !ioAllowedEgress(lname) {
                                d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_IO_EGRESS_UNSAFE", Message: "unsafe egress I/O operation: " + st.Name, File: path, Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}}
                                if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                            }
                        }
                    }

                    // Reachability and graph-based hints
                    nodes := map[string]bool{}
                    posByName := map[string]diag.Position{}
                    adj := map[string][]string{}
                    radj := map[string][]string{}
                    degree := map[string]int{}
                    foundIO := false
                    for _, s := range stmts {
                        switch n := s.(type) {
                        case *ast.StepStmt:
                            nodes[n.Name] = true
                            posByName[n.Name] = diag.Position{Line: n.Pos.Line, Column: n.Pos.Column, Offset: n.Pos.Offset}
                            if strings.HasPrefix(strings.ToLower(n.Name), "io.") { foundIO = true }
                        case *ast.EdgeStmt:
                            nodes[n.From] = true
                            nodes[n.To] = true
                            posByName[n.From] = diag.Position{Line: n.Pos.Line, Column: n.Pos.Column, Offset: n.Pos.Offset}
                            adj[n.From] = append(adj[n.From], n.To)
                            radj[n.To] = append(radj[n.To], n.From)
                            degree[n.From]++
                            degree[n.To]++
                        }
                    }
                    // BFS from ingress
                    vis := map[string]bool{}
                    q := []string{"ingress"}
                    for len(q) > 0 {
                        u := q[0]; q = q[1:]
                        if vis[u] { continue }
                        vis[u] = true
                        for _, v := range adj[u] { if !vis[v] { q = append(q, v) } }
                    }
                    // Reverse reachability from egress
                    visToEgress := map[string]bool{}
                    rq := []string{"egress"}
                    for len(rq) > 0 {
                        u := rq[0]; rq = rq[1:]
                        if visToEgress[u] { continue }
                        visToEgress[u] = true
                        for _, v := range radj[u] { if !visToEgress[v] { rq = append(rq, v) } }
                    }
                    // Non-fatal smells
                    for name := range nodes {
                        lower := strings.ToLower(name)
                        if lower != "ingress" && !vis[name] && degree[name] > 0 {
                            p := posByName[name]
                            d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_PIPELINE_UNREACHABLE_NODE", Message: "pipeline node appears unreachable from ingress", File: path, Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}, Data: map[string]any{"node": name}}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        }
                        if lower != "egress" && !visToEgress[name] && degree[name] > 0 {
                            p := posByName[name]
                            d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_PIPELINE_NONTERMINATING_NODE", Message: "pipeline node cannot reach egress", File: path, Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}, Data: map[string]any{"node": name}}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        }
                        if degree[name] == 0 && lower != "ingress" && lower != "egress" {
                            p := posByName[name]
                            d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_PIPELINE_DISCONNECTED_NODE", Message: "pipeline node has no incident edges", File: path, Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}, Data: map[string]any{"node": name}}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        }
                    }
                    // If no path from ingress to egress, surface a summary warning at pipeline name
                    if !vis["egress"] && nodes["ingress"] && nodes["egress"] {
                        d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_PIPELINE_NO_PATH_INGRESS_EGRESS", Message: "no path from ingress to egress; pipeline may not terminate", File: path, Pos: &diag.Position{Line: pd.NamePos.Line, Column: pd.NamePos.Column, Offset: pd.NamePos.Offset}}
                        if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                    }

                    // Capability/trust smells (non-fatal guidance)
                    if foundIO {
                        if !declCaps["io"] {
                            // pick a representative position
                            p := diag.Position{Line: pd.Pos.Line, Column: pd.Pos.Column, Offset: pd.Pos.Offset}
                            d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_CAPABILITY_UNDECLARED", Message: "io.* operations present; declare capability 'io' via #pragma capabilities", File: path, Pos: &p}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        }
                        if trustLevel == "" {
                            p := diag.Position{Line: pd.Pos.Line, Column: pd.Pos.Column, Offset: pd.Pos.Offset}
                            d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_TRUST_UNSPECIFIED", Message: "trust level unspecified; declare via #pragma trust level=trusted|untrusted", File: path, Pos: &p}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        }
                        if trustLevel == "untrusted" {
                            p := diag.Position{Line: pd.Pos.Line, Column: pd.Pos.Column, Offset: pd.Pos.Offset}
                            d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_TRUST_UNTRUSTED_IO", Message: "io.* under untrusted trust level; consider revising trust or removing I/O", File: path, Pos: &p}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        }
                    }
                    // Concurrency hints: limits unspecified/unused, schedule/policy smells
                    if len(nodes) > 0 {
                        if len(limits) == 0 {
                            d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_CONCURRENCY_LIMITS_UNSPECIFIED", Message: "concurrency limits unspecified; consider #pragma concurrency:limits", File: path, Pos: &diag.Position{Line: pd.NamePos.Line, Column: pd.NamePos.Column, Offset: pd.NamePos.Offset}}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        } else {
                            // count node kinds
                            kindCount := map[string]int{"ingress":0,"transform":0,"fanout":0,"collect":0,"mutable":0,"egress":0}
                            for name := range nodes {
                                ln := strings.ToLower(name)
                                if _, ok := kindCount[ln]; ok { kindCount[ln]++ }
                            }
                            for k := range limits {
                                if kindCount[k] == 0 {
                                    d := diag.Record{Timestamp: now, Level: diag.Info, Code: "W_CONCURRENCY_LIMIT_UNUSED", Message: "concurrency limit set for unused kind: " + k, File: path, Pos: &diag.Position{Line: pd.NamePos.Line, Column: pd.NamePos.Column, Offset: pd.NamePos.Offset}, Data: map[string]any{"kind": k}}
                                    if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                                }
                            }
                        }
                        if workers == 1 && (schedule == "worksteal" || schedule == "fair") {
                            d := diag.Record{Timestamp: now, Level: diag.Info, Code: "W_CONCURRENCY_SCHEDULE_IGNORED", Message: "concurrency schedule likely ineffective with workers=1", File: path, Pos: &diag.Position{Line: pd.NamePos.Line, Column: pd.NamePos.Column, Offset: pd.NamePos.Offset}, Data: map[string]any{"schedule": schedule, "workers": workers}}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        }
                        if workers > 1 && schedule == "" {
                            d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_CONCURRENCY_SCHEDULE_UNSPECIFIED", Message: "concurrency schedule unspecified; set via #pragma concurrency:schedule", File: path, Pos: &diag.Position{Line: pd.NamePos.Line, Column: pd.NamePos.Column, Offset: pd.NamePos.Offset}}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        }
                    }
                }
            }
            // RAII hint: release(x) should be wrapped: mutate(release(x))
            if t.StageB || t.RAIIHint {
                wrapped := collectMutateWrappedReleases(af)
                walkExprs(af, func(e ast.Expr) {
                    if ce, ok := e.(*ast.CallExpr); ok {
                        lname := strings.ToLower(ce.Name)
                        if strings.HasSuffix(lname, ".release") || lname == "release" {
                            key := callKey("", ce)
                            if !wrapped[key] {
                                d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_RAII_OWNED_HINT", Message: "wrap release in mutate(release(x)) for explicit handoff", File: path, Pos: &diag.Position{Line: ce.NamePos.Line, Column: ce.NamePos.Column, Offset: ce.NamePos.Offset}}
                                if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                            }
                        }
                    }
                })
            }
            return nil
        })
    }
    return out
}
