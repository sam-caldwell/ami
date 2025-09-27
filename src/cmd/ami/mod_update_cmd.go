package main

import "github.com/spf13/cobra"

// newModUpdateCmd creates `ami mod update`.
func newModUpdateCmd() *cobra.Command {
    var jsonOut bool
    cmd := &cobra.Command{
        Use:   "update",
        Short: "Update package cache from workspace packages and refresh ami.sum",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runModUpdate(cmd.OutOrStdout(), ".", jsonOut)
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-parsable JSON output")
    return cmd
}

