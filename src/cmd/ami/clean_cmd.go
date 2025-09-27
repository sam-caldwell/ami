package main

import (
    "os"
    "github.com/spf13/cobra"
)

// newCleanCmd creates the `ami clean` subcommand.
func newCleanCmd() *cobra.Command {
    var jsonOut bool
    cmd := &cobra.Command{
        Use:   "clean",
        Short: "Remove and recreate the build directory",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Resolve absolute working directory for deterministic behavior.
            wd, err := os.Getwd()
            if err != nil {
                return err
            }
            return runClean(cmd.OutOrStdout(), wd, jsonOut)
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-parsable JSON output")
    return cmd
}
