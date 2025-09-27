package main

import (
    "os"
    "github.com/spf13/cobra"
)

// newBuildCmd returns the `ami build` subcommand.
func newBuildCmd() *cobra.Command {
    var jsonOut bool
    cmd := &cobra.Command{
        Use:   "build",
        Short: "Validate workspace and build (phase: validation)",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Resolve absolute working directory for deterministic behavior.
            wd, err := os.Getwd()
            if err != nil { return err }
            verbose, _ := cmd.Flags().GetBool("verbose")
            return runBuild(cmd.OutOrStdout(), wd, jsonOut, verbose)
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON diagnostics and results")
    return cmd
}
