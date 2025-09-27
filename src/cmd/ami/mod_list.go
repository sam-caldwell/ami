package main

import "github.com/spf13/cobra"

// newModListCmd returns `ami mod list` subcommand.
func newModListCmd() *cobra.Command {
    var jsonOut bool
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List cached modules",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runModList(cmd.OutOrStdout(), jsonOut)
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-parsable JSON output")
    return cmd
}
