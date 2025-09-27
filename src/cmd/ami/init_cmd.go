package main

import (
    "os"
    "github.com/spf13/cobra"
)

// newInitCmd creates the `ami init` subcommand.
func newInitCmd() *cobra.Command {
    var force bool
    var jsonOut bool
    cmd := &cobra.Command{
        Use:   "init",
        Short: "Initialize an AMI workspace",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Resolve absolute working directory to avoid surprises with relative paths.
            wd, err := os.Getwd()
            if err != nil {
                return err
            }
            return runInit(cmd.OutOrStdout(), wd, force, jsonOut)
        },
    }
    cmd.Flags().BoolVar(&force, "force", false, "force initialization (fill missing fields; git init if needed)")
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-parsable JSON output")
    return cmd
}
