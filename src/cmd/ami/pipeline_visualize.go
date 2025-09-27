package main

import (
    "encoding/json"
    "github.com/spf13/cobra"
    "github.com/sam-caldwell/ami/src/schemas/graph"
)

// newPipelineVisualizeCmd returns `ami pipeline visualize` subcommand.
// Currently emits a placeholder graph in JSON mode; ASCII renderer TBD.
func newPipelineVisualizeCmd() *cobra.Command {
    var jsonOut bool
    cmd := &cobra.Command{
        Use:   "visualize",
        Short: "Render ASCII pipeline graphs (stub)",
        RunE: func(cmd *cobra.Command, args []string) error {
            if jsonOut {
                g := graph.Graph{
                    Package: "main",
                    Unit:    "",
                    Name:    "Placeholder",
                    Nodes: []graph.Node{
                        {ID: "ingress", Kind: "ingress", Label: "ingress"},
                        {ID: "worker", Kind: "worker", Label: "worker"},
                        {ID: "egress",  Kind: "egress",  Label: "egress"},
                    },
                    Edges: []graph.Edge{
                        {From: "ingress", To: "worker"},
                        {From: "worker",  To: "egress"},
                    },
                }
                return json.NewEncoder(cmd.OutOrStdout()).Encode(g)
            }
            // ASCII renderer to be implemented later; show help for now.
            return cmd.Help()
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-parsable JSON output (graph.v1)")
    return cmd
}
