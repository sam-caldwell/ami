package main

import (
    "github.com/spf13/cobra"
)

// newPipelineVisualizeCmd returns `ami pipeline visualize` subcommand (stub).
func newPipelineVisualizeCmd() *cobra.Command {
    var jsonOut bool
    cmd := &cobra.Command{
        Use:   "visualize",
        Short: "Render ASCII pipeline graphs (stub)",
        RunE: func(cmd *cobra.Command, args []string) error {
            // For now, just show help to avoid side effects.
            return cmd.Help()
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-parsable JSON output (future)")
    return cmd
}

