package main

import "github.com/spf13/cobra"

// newModCleanCmd creates `ami mod clean`.
func newModCleanCmd() *cobra.Command {
    var jsonOut bool
    cmd := &cobra.Command{
        Use:   "clean",
        Short: "Remove and recreate the AMI package cache",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runModClean(cmd.OutOrStdout(), jsonOut)
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-parsable JSON output")
    return cmd
}

