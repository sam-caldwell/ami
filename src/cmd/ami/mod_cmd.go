package main

import "github.com/spf13/cobra"

// newModCmd is the parent `ami mod` command to group module operations.
func newModCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "mod",
        Short: "Module cache operations",
        RunE: func(cmd *cobra.Command, args []string) error { return cmd.Help() },
    }
    cmd.AddCommand(newModCleanCmd())
    cmd.AddCommand(newModListCmd())
    cmd.AddCommand(newModUpdateCmd())
    cmd.AddCommand(newModSumCmd())
    cmd.AddCommand(newModGetCmd())
    cmd.AddCommand(newModAuditCmd())
    return cmd
}
