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

type sourceWithPos struct{ line, col int }

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
                semDiags := append(sem.AnalyzePipelineSemantics(af), sem.AnalyzeErrorSemantics(af)...)
                semDiags = append(semDiags, sem.AnalyzeDecorators(af)...)
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
                        key := im.Alias
                        if prior, dup := seenAlias[key]; dup {
                            d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_DUP_IMPORT_ALIAS", Message: "duplicate import alias: " + key, File: path, Pos: &diag.Position{Line: im.AliasPos.Line, Column: im.AliasPos.Column, Offset: im.AliasPos.Offset}, Data: map[string]any{"prevLine": prior.line, "prevColumn": prior.col}}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        } else {
                            seenAlias[key] = sourceWithPos{line: im.AliasPos.Line, col: im.AliasPos.Column}
                        }
                    }
                    if im.Path != "" && !strings.ContainsAny(im.Path, "/\".") { // looks like bare ident
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
                // Gather declared capabilities and trust level from pragmas (parser + raw scan fallback)
                declCaps := map[string]bool{}
                trustLevel := ""
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
                        }
                    }
                }
                for _, dcl := range af.Decls {
                    pd, ok := dcl.(*ast.PipelineDecl); if !ok || pd == nil { continue }
                    var stmts []ast.Stmt
                    if len(pd.Stmts) > 0 { stmts = pd.Stmts } else if pd.Body != nil { stmts = pd.Body.Stmts }
                    // Backpressure hints: scan step call names and attributes
                    for _, s := range stmts {
                        st, ok := s.(*ast.StepStmt); if !ok { continue }
                        // Capability (I/O) check: forbid io.* nodes outside ingress/egress
                        lname := strings.ToLower(st.Name)
                        if strings.HasPrefix(lname, "io.") && lname != "ingress" && lname != "egress" {
                            d := diag.Record{Timestamp: now, Level: diag.Error, Code: "E_IO_PERMISSION", Message: "io.* operations only allowed in ingress/egress nodes", File: path, Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}}
                            if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                        }
                        // Capability declaration required for io.* even when positionally allowed
                        if strings.HasPrefix(lname, "io.") {
                            // derive specific capability
                            cap := "io"
                            // map common io verbs
                            if strings.HasPrefix(lname, "io.read") || strings.HasPrefix(lname, "io.recv") { cap = "io.read" }
                            if strings.HasPrefix(lname, "io.write") || strings.HasPrefix(lname, "io.send") { cap = "io.write" }
                            if strings.HasPrefix(lname, "io.connect") || strings.HasPrefix(lname, "io.listen") || strings.HasPrefix(lname, "io.dial") { cap = "network" }
                            // allow generic io capability for read/write specifics
                            allowed := declCaps[cap] || declCaps["io"]
                            if !allowed {
                                d := diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CAPABILITY_REQUIRED", Message: "operation requires capability '" + cap + "'", File: path, Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}, Data: map[string]any{"cap": cap}}
                                if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                            }
                            // trust enforcement: forbid network for untrusted
                            if (trustLevel == "untrusted" || trustLevel == "low") && cap == "network" {
                                d := diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TRUST_VIOLATION", Message: "operation not allowed under trust level '" + trustLevel + "'", File: path, Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}, Data: map[string]any{"trust": trustLevel}}
                                if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                            }
                        }
                        // Step-call form: merge.Buffer(capacity, policy)
                        if strings.HasSuffix(lname, ".buffer") {
                            capVal := 0
                            if len(st.Args) >= 1 { capVal = atoiSafe(st.Args[0].Text) }
                            var policy string
                            if len(st.Args) >= 2 { policy = strings.ToLower(st.Args[1].Text) }
                            if policy == "drop" {
                                d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_PIPELINE_BUFFER_DROP_ALIAS", Message: "ambiguous backpressure alias 'drop'; use dropOldest/dropNewest/block", File: path, Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}}
                                if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                            }
                            if capVal <= 1 && (policy == "dropoldest" || policy == "dropnewest") {
                                d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_PIPELINE_BUFFER_POLICY_SMELL", Message: "buffer policy with capacity<=1 is likely ineffective", File: path, Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}, Data: map[string]any{"capacity": capVal, "policy": policy}}
                                if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                            }
                        }
                        // Also inspect raw args for attribute-like calls (e.g., Collect(merge.Sort(...)))
                        for _, a := range st.Args {
                            txt := strings.TrimSpace(strings.ToLower(a.Text))
                            // merge.Sort(field[,order]) in argument form
                            if strings.HasPrefix(txt, "merge.sort(") {
                                // extract inside parens
                                inner := strings.TrimSuffix(strings.TrimPrefix(txt, "merge.sort("), ")")
                                // split by comma into up to two parts
                                parts := []string{}
                                for _, p := range strings.Split(inner, ",") {
                                    p = strings.TrimSpace(p)
                                    if p != "" { parts = append(parts, p) }
                                }
                                if len(parts) == 0 {
                                    d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_SORT_NO_FIELD", Message: "merge.Sort requires a field argument", File: path, Pos: &diag.Position{Line: a.Pos.Line, Column: a.Pos.Column, Offset: a.Pos.Offset}}
                                    if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                                } else if len(parts) >= 2 {
                                    ord := parts[1]
                                    if ord != "asc" && ord != "desc" {
                                        d := diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_SORT_ORDER_INVALID", Message: "merge.Sort order must be 'asc' or 'desc'", File: path, Pos: &diag.Position{Line: a.Pos.Line, Column: a.Pos.Column, Offset: a.Pos.Offset}, Data: map[string]any{"order": parts[1]}}
                                        if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                                    }
                                }
                            }
                        }
                        for _, a := range st.Attrs {
                            name := strings.ToLower(a.Name)
                            if name == "buffer" || strings.HasSuffix(name, ".buffer") {
                                // capacity is first arg
                                capVal := 0
                                if len(a.Args) >= 1 { capVal = atoiSafe(a.Args[0].Text) }
                                var policy string
                                if len(a.Args) >= 2 { policy = strings.ToLower(a.Args[1].Text) }
                                if policy == "drop" {
                                    d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_PIPELINE_BUFFER_DROP_ALIAS", Message: "ambiguous backpressure alias 'drop'; use dropOldest/dropNewest/block", File: path, Pos: &diag.Position{Line: a.Pos.Line, Column: a.Pos.Column, Offset: a.Pos.Offset}}
                                    if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                                }
                                if capVal <= 1 && (policy == "dropoldest" || policy == "dropnewest") {
                                    d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_PIPELINE_BUFFER_POLICY_SMELL", Message: "buffer policy with capacity<=1 is likely ineffective", File: path, Pos: &diag.Position{Line: a.Pos.Line, Column: a.Pos.Column, Offset: a.Pos.Offset}, Data: map[string]any{"capacity": capVal, "policy": policy}}
                                    if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                                }
                            }
                            if strings.HasPrefix(name, "io.") {
                                cap := "io"
                                if strings.HasPrefix(name, "io.read") || strings.HasPrefix(name, "io.recv") { cap = "io.read" }
                                if strings.HasPrefix(name, "io.write") || strings.HasPrefix(name, "io.send") { cap = "io.write" }
                                if strings.HasPrefix(name, "io.connect") || strings.HasPrefix(name, "io.listen") || strings.HasPrefix(name, "io.dial") { cap = "network" }
                                allowed := declCaps[cap] || declCaps["io"]
                            if !allowed {
                                d := diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CAPABILITY_REQUIRED", Message: "operation requires capability '" + cap + "'", File: path, Pos: &diag.Position{Line: a.Pos.Line, Column: a.Pos.Column, Offset: a.Pos.Offset}, Data: map[string]any{"cap": cap}}
                                if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                            }
                            if (trustLevel == "untrusted" || trustLevel == "low") && cap == "network" {
                                d := diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TRUST_VIOLATION", Message: "operation not allowed under trust level '" + trustLevel + "'", File: path, Pos: &diag.Position{Line: a.Pos.Line, Column: a.Pos.Column, Offset: a.Pos.Offset}, Data: map[string]any{"trust": trustLevel}}
                                if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                            }
                            }
                            // merge.Sort(field[,order]) validation (attribute form)
                            if strings.HasSuffix(name, ".sort") {
                                if len(a.Args) == 0 || a.Args[0].Text == "" {
                                    d := diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_SORT_NO_FIELD", Message: "merge.Sort requires a field argument", File: path, Pos: &diag.Position{Line: a.Pos.Line, Column: a.Pos.Column, Offset: a.Pos.Offset}}
                                    if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                                }
                                if len(a.Args) >= 2 {
                                    ord := strings.ToLower(a.Args[1].Text)
                                    if ord != "asc" && ord != "desc" {
                                        d := diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_SORT_ORDER_INVALID", Message: "merge.Sort order must be 'asc' or 'desc'", File: path, Pos: &diag.Position{Line: a.Pos.Line, Column: a.Pos.Column, Offset: a.Pos.Offset}, Data: map[string]any{"order": a.Args[1].Text}}
                                        if m := disables[path]; m == nil || !m[d.Code] { out = append(out, d) }
                                    }
                                }
                            }
                        }
                    }
                    // Reachability and connectivity over explicit edge statements
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

// collectUsedIdents returns a set of identifier names used as top-level prefixes in expressions.
func collectUsedIdents(f *ast.File) map[string]bool {
    used := map[string]bool{}
    walkExprs(f, func(e ast.Expr) {
        switch n := e.(type) {
        case *ast.IdentExpr:
            used[n.Name] = true
        case *ast.CallExpr:
            // Split qualified name on '.' and take first segment
            name := n.Name
            if i := strings.IndexByte(name, '.'); i >= 0 { name = name[:i] }
            used[name] = true
        }
    })
    return used
}

// walkStmts invokes fn for every statement in function and pipeline bodies.
func walkStmts(f *ast.File, fn func(ast.Stmt)) {
    for _, d := range f.Decls {
        switch n := d.(type) {
        case *ast.FuncDecl:
            if n != nil && n.Body != nil {
                for _, s := range n.Body.Stmts { fn(s) }
            }
        case *ast.PipelineDecl:
            if n != nil && n.Body != nil {
                for _, s := range n.Body.Stmts { fn(s) }
            }
        }
    }
}

// walkExprs invokes fn for every expression node reachable from functions/pipelines.
func walkExprs(f *ast.File, fn func(ast.Expr)) {
    walkStmts(f, func(s ast.Stmt) {
        switch n := s.(type) {
        case *ast.ExprStmt:
            if n.X != nil { walkExpr(n.X, fn) }
        case *ast.AssignStmt:
            if n.Value != nil { walkExpr(n.Value, fn) }
        case *ast.VarDecl:
            if n.Init != nil { walkExpr(n.Init, fn) }
        case *ast.DeferStmt:
            if n.Call != nil { walkExpr(n.Call, fn) }
        case *ast.ReturnStmt:
            for _, e := range n.Results { walkExpr(e, fn) }
        }
    })
}

func walkExpr(e ast.Expr, fn func(ast.Expr)) {
    if e == nil { return }
    fn(e)
    switch n := e.(type) {
    case *ast.CallExpr:
        for _, a := range n.Args { walkExpr(a, fn) }
    case *ast.BinaryExpr:
        if n.X != nil { walkExpr(n.X, fn) }
        if n.Y != nil { walkExpr(n.Y, fn) }
    case *ast.SelectorExpr:
        if n.X != nil { walkExpr(n.X, fn) }
    }
}

// collectMutateWrappedReleases returns keys of release calls that are wrapped in mutate().
func collectMutateWrappedReleases(f *ast.File) map[string]bool {
    wrapped := map[string]bool{}
    walkExprs(f, func(e ast.Expr) {
        ce, ok := e.(*ast.CallExpr)
        if !ok { return }
        if strings.EqualFold(ce.Name, "mutate") && len(ce.Args) > 0 {
            if inner, ok := ce.Args[0].(*ast.CallExpr); ok {
                lname := strings.ToLower(inner.Name)
                if lname == "release" || strings.HasSuffix(lname, ".release") {
                    wrapped[callKey("", inner)] = true
                }
            }
        }
    })
    return wrapped
}

func callKey(prefix string, ce *ast.CallExpr) string {
    if ce == nil { return "" }
    return prefix + "@" + ce.Name + ":" + itoa(ce.NamePos.Line) + ":" + itoa(ce.NamePos.Column)
}

func itoa(i int) string { return intToString(i) }

func intToString(i int) string {
    // fast small-int conversion without fmt
    if i == 0 { return "0" }
    neg := i < 0
    if neg { i = -i }
    var b [20]byte
    p := len(b)
    for i > 0 {
        p--
        b[p] = byte('0' + (i % 10))
        i /= 10
    }
    if neg { p--; b[p] = '-' }
    return string(b[p:])
}

func atoiSafe(s string) int {
    // simple unsigned parse; non-numeric returns 0
    n := 0
    for i := 0; i < len(s); i++ {
        c := s[i]
        if c < '0' || c > '9' { break }
        n = n*10 + int(c-'0')
        if n > 1_000_000_000 { break }
    }
    return n
}
