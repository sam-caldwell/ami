package main

import (
    "encoding/json"
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
            // honor global --json flag
            if f := cmd.Flags().Lookup("json"); f != nil {
                useJSON := false
                if v, err := cmd.Flags().GetBool("json"); err == nil { useJSON = v }
                if useJSON {
                    _ = json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{"version": version})
                    return nil
                }
            }
            _, _ = fmt.Fprintln(cmd.OutOrStdout(), "ami version "+version)
            return nil
        },
    }
}
