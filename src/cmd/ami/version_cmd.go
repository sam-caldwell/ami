package main

import (
    "fmt"
    "github.com/spf13/cobra"
)

// version holds the current CLI version.
// It is intended to be overridden at build time via:
//   go build -ldflags "-X main.version=v1.2.3"
// Defaults to a dev marker when not injected.
var version = "v0.0.0-dev"

// newVersionCmd returns the `ami version` subcommand.
func newVersionCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "version",
        Short: "Print version information",
        RunE: func(cmd *cobra.Command, args []string) error {
            _, _ = fmt.Fprintln(cmd.OutOrStdout(), "ami version "+version)
            return nil
        },
    }
}
