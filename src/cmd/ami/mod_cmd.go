package main

import "github.com/spf13/cobra"

// newModCmd is the parent `ami mod` command to group module operations.
func newModCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "mod",
        Short: "Module cache operations",
        Example: "\n  # List packages in the cache\n  ami mod list\n\n  # Fetch a local package into the cache and update ami.sum\n  ami mod get ./vendor/alpha\n\n  # Validate ami.sum against cache contents\n  ami mod sum --json\n\n  # Update cache from workspace packages and rewrite ami.sum\n  ami mod update --json\n\n  # Clean (reset) the package cache\n  ami mod clean --json\n",
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
