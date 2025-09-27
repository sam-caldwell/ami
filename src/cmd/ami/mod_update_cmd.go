package main

import "github.com/spf13/cobra"

// newModUpdateCmd creates `ami mod update`.
func newModUpdateCmd() *cobra.Command {
    var jsonOut bool
    cmd := &cobra.Command{
        Use:   "update",
        Short: "Update package cache from workspace packages and refresh ami.sum",
        Example: "\n  # Copy workspace packages into the cache and rewrite ami.sum\n  ami mod update\n\n  # JSON output for CI logs\n  ami mod update --json\n",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runModUpdate(cmd.OutOrStdout(), ".", jsonOut)
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-parsable JSON output")
    return cmd
}
