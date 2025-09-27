package main

import (
    "fmt"
    "github.com/spf13/cobra"
)

// versionString holds the current CLI version.
// TODO: replace with build-time injection in later milestones.
const versionString = "v0.0.0-dev"

// newVersionCmd returns the `ami version` subcommand.
func newVersionCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "version",
        Short: "Print version information",
        RunE: func(cmd *cobra.Command, args []string) error {
            _, _ = fmt.Fprintln(cmd.OutOrStdout(), "ami version "+versionString)
            return nil
        },
    }
}

