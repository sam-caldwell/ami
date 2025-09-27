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
                enc := json.NewEncoder(cmd.OutOrStdout())
                for _, g := range graphs { _ = enc.Encode(g) }
                // summary object
                sum := map[string]any{"schema": graph.Schema, "type": "summary", "pipelines": len(graphs)}
                _ = enc.Encode(sum)
                return nil
            }
            // Human: header + one-line ASCII per pipeline
            for i, g := range graphs {
                header := fmt.Sprintf("package: %s  pipeline: %s\n", g.Package, g.Name)
                line := ascii.RenderLine(g, ascii.Options{})
                if i > 0 { _, _ = cmd.OutOrStdout().Write([]byte("\n")) }
                _, _ = cmd.OutOrStdout().Write([]byte(header))
                _, _ = cmd.OutOrStdout().Write([]byte(line + "\n"))
            }
            return nil
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-parsable JSON output (graph.v1)")
    cmd.Flags().StringVar(&pkgKey, "package", "", "visualize a specific workspace package key (e.g., main)")
    cmd.Flags().StringVar(&fileOnly, "file", "", "visualize only a specific .ami file path")
    return cmd
}

// graphFromPipeline constructs a simple straight-line graph from a PipelineDecl's step order.
func graphFromPipeline(pkg string, unit string, pd *ast.PipelineDecl) graph.Graph {
    g := graph.Graph{Package: pkg, Unit: unit, Name: pd.Name}
    // Collect steps in order
    var ids []string
    for i, s := range pd.Stmts {
        st, ok := s.(*ast.StepStmt); if !ok { continue }
        // normalize kind: lower-case last segment of Name
        parts := strings.Split(st.Name, ".")
        kind := strings.ToLower(parts[len(parts)-1])
        id := fmt.Sprintf("%02d:%s", i, kind)
        lbl := kind
        g.Nodes = append(g.Nodes, graph.Node{ID: id, Kind: kind, Label: lbl})
        ids = append(ids, id)
    }
    // Sequential edges
    for i := 0; i+1 < len(ids); i++ {
        g.Edges = append(g.Edges, graph.Edge{From: ids[i], To: ids[i+1]})
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
