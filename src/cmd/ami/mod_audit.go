package main

import (
    "os"
    "github.com/spf13/cobra"
)

// newModAuditCmd returns `ami mod audit` subcommand.
func newModAuditCmd() *cobra.Command {
    var jsonOut bool
    cmd := &cobra.Command{
        Use:   "audit",
        Short: "Audit workspace dependencies vs ami.sum and cache",
        RunE: func(cmd *cobra.Command, args []string) error {
            wd, err := os.Getwd()
            if err != nil { return err }
            return runModAudit(cmd.OutOrStdout(), wd, jsonOut)
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-parsable JSON output")
    return cmd
}

