package main

import "github.com/spf13/cobra"

// newPipelineCmd is the parent `ami pipeline` command to group pipeline operations.
func newPipelineCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "pipeline",
        Short: "Pipeline utilities",
        Example: "\n  # Visualize pipelines as ASCII\n  ami pipeline visualize\n\n  # JSON graph output (graph.v1)\n  ami pipeline visualize --json\n",
        RunE: func(cmd *cobra.Command, args []string) error { return cmd.Help() },
    }
    // Subcommands
    cmd.AddCommand(newPipelineVisualizeCmd())
    cmd.AddCommand(newPipelineStatsCmd())
    return cmd
}
