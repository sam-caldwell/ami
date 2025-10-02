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
    var showIDs bool
    var edgeIDs bool
    cmd := &cobra.Command{
        Use:   "visualize",
        Short: "Render ASCII pipeline graphs",
        Example: "\n  # Human ASCII output\n  ami pipeline visualize\n\n  # Focus on nodes matching substring\n  ami pipeline visualize --focus egress\n\n  # Visualize only a specific file\n  ami pipeline visualize --file src/main.ami\n\n  # JSON output and omit summary record\n  ami pipeline visualize --json --no-summary\n",
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
            if pkgKey != "" { pkg = findPackageByRootKey(&ws, pkgKey) }
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
                b, rerr := os.ReadFile(fpath); if rerr != nil { continue }
                sf := &source.File{Name: fpath, Content: string(b)}
                p := parser.New(sf)
                af, _ := p.ParseFile()
                unit := filepath.Base(fpath)
                for _, d := range af.Decls {
                    if pd, ok := d.(*ast.PipelineDecl); ok {
                        g := graphFromPipeline(pkg.Name, unit, pd)
                        // Detect circular references early; emit diagnostics and abort
                        if hasCycle, cyc := detectCycle(g); hasCycle {
                            if jsonOut {
                                rec := diag.Record{Level: diag.Error, Code: "E_GRAPH_CYCLE", Message: "circular reference detected", Package: g.Package, File: fpath, Data: map[string]any{"pipeline": g.Name, "cycle": cyc}}
                                _ = json.NewEncoder(cmd.OutOrStdout()).Encode(rec)
                            }
                            return fmt.Errorf("circular reference detected in %s", fpath)
                        }
                        graphs = append(graphs, g)
                    }
                }
            }
            focus, _ := cmd.Flags().GetString("focus")
            width, _ := cmd.Flags().GetInt("width")
            noSummary, _ := cmd.Flags().GetBool("no-summary")
            excludes, _ := cmd.Flags().GetStringSlice("json-exclude")
            legend, _ := cmd.Flags().GetBool("legend")
            color, _ := cmd.Flags().GetBool("color")
            if jsonOut {
                // Filter by focus if provided
                if focus != "" {
                    lf := strings.ToLower(focus)
                    tmp := graphs[:0]
                    for _, g := range graphs { if graphContains(g, lf) { tmp = append(tmp, g) } }
                    graphs = tmp
                }
                // Exclude fields if requested
                if len(excludes) > 0 {
                    for i := range graphs { graphs[i] = applyJSONExcludes(graphs[i], excludes) }
                }
                enc := json.NewEncoder(cmd.OutOrStdout())
                for _, g := range graphs { _ = enc.Encode(g) }
                if !noSummary { _ = enc.Encode(map[string]any{"schema": "graph.batch.v1", "type": "summary", "count": len(graphs)}) }
                return nil
            }
            // ASCII output
            for i, g := range graphs {
                header, line := ascii.Render(g, ascii.Options{Width: width, Focus: focus, Legend: legend, Color: color, ShowIDs: showIDs, EdgeIDs: edgeIDs})
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
    cmd.Flags().StringSlice("json-exclude", nil, "comma-separated fields to exclude in JSON (e.g., attrs)")
    cmd.Flags().Bool("legend", false, "show a simple legend in ASCII output")
    cmd.Flags().BoolVar(&showIDs, "show-ids", false, "prefix node labels with instance ids (e.g., 00:kind)")
    cmd.Flags().BoolVar(&edgeIDs, "edge-ids", false, "label edges by fromId->toId in ASCII output")
    return cmd
}
