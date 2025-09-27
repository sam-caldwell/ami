package main

import "github.com/spf13/cobra"

// newPipelineCmd is the parent `ami pipeline` command to group pipeline operations.
func newPipelineCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "pipeline",
        Short: "Pipeline utilities",
        RunE: func(cmd *cobra.Command, args []string) error { return cmd.Help() },
    }
    return cmd
}

