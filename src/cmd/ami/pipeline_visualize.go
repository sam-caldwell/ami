package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"

    "github.com/spf13/cobra"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/visualize/ascii"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/schemas/diag"
    "github.com/sam-caldwell/ami/src/schemas/graph"
)

// newPipelineVisualizeCmd returns `ami pipeline visualize` subcommand.
// Currently emits a placeholder graph in JSON mode; ASCII renderer TBD.
func newPipelineVisualizeCmd() *cobra.Command {
    var jsonOut bool
    var pkgKey string
    var fileOnly string
    cmd := &cobra.Command{
        Use:   "visualize",
        Short: "Render ASCII pipeline graphs",
        RunE: func(cmd *cobra.Command, args []string) error {
            dir := "."
            wsPath := filepath.Join(dir, "ami.workspace")
            var ws workspace.Workspace
            if _, err := os.Stat(wsPath); errors.Is(err, os.ErrNotExist) {
                if jsonOut {
                    rec := diag.Record{Level: diag.Error, Code: "E_WS_SCHEMA", Message: "workspace not found", File: "ami.workspace"}
                    _ = json.NewEncoder(cmd.OutOrStdout()).Encode(rec)
                    return nil
                }
                return fmt.Errorf("workspace not found: ami.workspace")
            } else if err != nil {
                return fmt.Errorf("stat workspace: %v", err)
            }
            if err := ws.Load(wsPath); err != nil {
                if jsonOut {
                    rec := diag.Record{Level: diag.Error, Code: "E_WS_SCHEMA", Message: "failed to load workspace: " + err.Error(), File: "ami.workspace"}
                    _ = json.NewEncoder(cmd.OutOrStdout()).Encode(rec)
                    return nil
                }
                return fmt.Errorf("failed to load workspace: %v", err)
            }
            // Determine package to visualize
            pkg := ws.FindPackage("main")
            if pkgKey != "" {
                pkg = findPackageByRootKey(&ws, pkgKey)
            }
            if pkg == nil {
                if jsonOut {
                    rec := diag.Record{Level: diag.Error, Code: "E_WS_SCHEMA", Message: "missing main package", File: "ami.workspace"}
                    _ = json.NewEncoder(cmd.OutOrStdout()).Encode(rec)
                    return nil
                }
                return fmt.Errorf("missing main package in workspace")
            }
            root := filepath.Clean(filepath.Join(dir, pkg.Root))
            // Discover .ami files under root (optionally filter by --file)
            var files []string
            err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
                if err != nil { return nil }
                if d.IsDir() { return nil }
                if filepath.Ext(path) != ".ami" { return nil }
                if fileOnly != "" {
                    // allow relative or absolute matches
                    ap, _ := filepath.Abs(path)
                    af, _ := filepath.Abs(fileOnly)
                    rp, _ := filepath.Rel(dir, path)
                    if !(path == fileOnly || ap == af || rp == fileOnly) { return nil }
                }
                files = append(files, path)
                return nil
            })
            if err != nil { return fmt.Errorf("walk: %v", err) }
            sort.Strings(files)
            // Parse pipelines and emit
            var graphs []graph.Graph
            for _, fpath := range files {
                b, rerr := os.ReadFile(fpath)
                if rerr != nil { continue }
                sf := &source.File{Name: fpath, Content: string(b)}
                p := parser.New(sf)
                af, _ := p.ParseFile()
                unit := filepath.Base(fpath)
                for _, d := range af.Decls {
                    if pd, ok := d.(*ast.PipelineDecl); ok {
                        g := graphFromPipeline(pkg.Name, unit, pd)
                        graphs = append(graphs, g)
                    }
                }
            }
            if jsonOut {
                noSummary, _ := cmd.Flags().GetBool("no-summary")
                enc := json.NewEncoder(cmd.OutOrStdout())
                for _, g := range graphs { _ = enc.Encode(g) }
                // summary object
                if !noSummary {
                    sum := map[string]any{"schema": graph.Schema, "type": "summary", "pipelines": len(graphs)}
                    _ = enc.Encode(sum)
                }
                return nil
            }
            // Optional focus filter for human mode
            focus, _ := cmd.Flags().GetString("focus")
            if focus != "" {
                ff := strings.ToLower(focus)
                var filtered []graph.Graph
                for _, g := range graphs {
                    if graphContains(g, ff) { filtered = append(filtered, g) }
                }
                graphs = filtered
            }
            // Human: header + ASCII block per pipeline
            width, _ := cmd.Flags().GetInt("width")
            for i, g := range graphs {
                header := fmt.Sprintf("package: %s  pipeline: %s\n", g.Package, g.Name)
                line := ascii.RenderBlock(g, ascii.Options{Width: width})
                if i > 0 { _, _ = cmd.OutOrStdout().Write([]byte("\n")) }
                _, _ = cmd.OutOrStdout().Write([]byte(header))
                _, _ = cmd.OutOrStdout().Write([]byte(line))
            }
            return nil
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-parsable JSON output (graph.v1)")
    cmd.Flags().StringVar(&pkgKey, "package", "", "visualize a specific workspace package key (e.g., main)")
    cmd.Flags().StringVar(&fileOnly, "file", "", "visualize only a specific .ami file path")
    cmd.Flags().String("focus", "", "only show pipelines that include this node substring")
    cmd.Flags().Int("width", 0, "wrap ASCII lines to this width (0=disable)")
    cmd.Flags().Bool("no-summary", false, "omit JSON summary record")
    return cmd
}

// graphFromPipeline constructs a simple straight-line graph from a PipelineDecl's step order.
func graphFromPipeline(pkg string, unit string, pd *ast.PipelineDecl) graph.Graph {
    g := graph.Graph{Package: pkg, Unit: unit, Name: pd.Name}
    // Collect steps and map names to node IDs
    var ids []string
    nameToID := map[string]string{}
    for i, s := range pd.Stmts {
        st, ok := s.(*ast.StepStmt)
        if !ok { continue }
        parts := strings.Split(st.Name, ".")
        base := parts[len(parts)-1]
        kind := strings.ToLower(base)
        label := base
        for _, at := range st.Attrs {
            if (at.Name == "type" || at.Name == "Type") && len(at.Args) > 0 {
                if at.Args[0].Text != "" { label = base + ":" + at.Args[0].Text }
            }
        }
        id := fmt.Sprintf("%02d:%s", i, strings.ToLower(base))
        g.Nodes = append(g.Nodes, graph.Node{ID: id, Kind: kind, Label: strings.ToLower(label)})
        ids = append(ids, id)
        if _, ok := nameToID[st.Name]; !ok { nameToID[st.Name] = id }
    }
    // Add explicit edges when present; otherwise chain sequentially
    var hasExplicit bool
    for _, s := range pd.Stmts { if _, ok := s.(*ast.EdgeStmt); ok { hasExplicit = true; break } }
    if hasExplicit {
        for _, s := range pd.Stmts {
            if e, ok := s.(*ast.EdgeStmt); ok {
                if fromID, ok1 := nameToID[e.From]; ok1 {
                    if toID, ok2 := nameToID[e.To]; ok2 {
                        // Annotate with attributes derived from the source step (fromID)
                        attrs := deriveEdgeAttrs(pd, fromID, nameToID)
                        g.Edges = append(g.Edges, graph.Edge{From: fromID, To: toID, Attrs: attrs})
                    }
                }
            }
        }
    } else {
        for i := 0; i+1 < len(ids); i++ {
            attrs := deriveEdgeAttrs(pd, ids[i], nameToID)
            g.Edges = append(g.Edges, graph.Edge{From: ids[i], To: ids[i+1], Attrs: attrs})
        }
    }
    return g
}

// findPackageByRootKey returns a package by the PackageList key (e.g., "main").
func findPackageByRootKey(ws *workspace.Workspace, key string) *workspace.Package {
    for i := range ws.Packages {
        if ws.Packages[i].Key == key { return &ws.Packages[i].Package }
    }
    return nil
}

// graphContains checks if g's name or any node label/kind contains focus (lowercase substring).
func graphContains(g graph.Graph, focus string) bool {
    if strings.Contains(strings.ToLower(g.Name), focus) { return true }
    for _, n := range g.Nodes {
        if strings.Contains(strings.ToLower(n.Label), focus) || strings.Contains(strings.ToLower(n.Kind), focus) {
            return true
        }
    }
    return false
}

// deriveEdgeAttrs inspects the step corresponding to fromID and returns edge attrs.
// Recognizes:
// - merge.Buffer(size, policy) → bounded (size>0), delivery (policy→bestEffort for dropOldest/dropNewest; atLeastOnce for block)
// - type(TypeName) → type
func deriveEdgeAttrs(pd *ast.PipelineDecl, fromID string, nameToID map[string]string) map[string]any {
    // find the step index matching fromID by scanning the constructed IDs
    // fromID is of the form "NN:kind" where NN is zero-padded index.
    // Extract index safely.
    var idx int
    if len(fromID) >= 2 && fromID[2] == ':' {
        // parse first two digits
        tens := int(fromID[0]-'0')
        ones := int(fromID[1]-'0')
        if 0 <= tens && tens <= 9 && 0 <= ones && ones <= 9 {
            idx = tens*10 + ones
        }
    }
    // count StepStmt to match idx
    si := -1
    for i, s := range pd.Stmts {
        if _, ok := s.(*ast.StepStmt); ok { si++; if si == idx { return attrsFromStep(s.(*ast.StepStmt)) } }
    }
    return nil
}

func attrsFromStep(st *ast.StepStmt) map[string]any {
    if st == nil { return nil }
    var bounded bool
    delivery := ""
    typ := ""
    for _, at := range st.Attrs {
        if (at.Name == "type" || at.Name == "Type") && len(at.Args) > 0 {
            typ = at.Args[0].Text
        }
        if len(at.Name) >= 6 && at.Name[:6] == "merge." {
            if at.Name == "merge.Buffer" {
                // args: size, policy
                if len(at.Args) > 0 {
                    if at.Args[0].Text != "0" && at.Args[0].Text != "" { bounded = true }
                }
                if len(at.Args) > 1 {
                    pol := at.Args[1].Text
                    switch pol {
                    case "dropOldest", "dropNewest":
                        delivery = "bestEffort"
                    case "block":
                        delivery = "atLeastOnce"
                    }
                }
            }
        }
    }
    m := map[string]any{}
    if bounded { m["bounded"] = true }
    if delivery != "" { m["delivery"] = delivery }
    if typ != "" { m["type"] = typ }
    if len(m) == 0 { return nil }
    return m
}
