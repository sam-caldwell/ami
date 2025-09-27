package main

import "github.com/spf13/cobra"

// newModSumCmd returns `ami mod sum` subcommand.
func newModSumCmd() *cobra.Command {
    var jsonOut bool
    cmd := &cobra.Command{
        Use:   "sum",
        Short: "Validate ami.sum against basic schema",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runModSum(cmd.OutOrStdout(), ".", jsonOut)
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-parsable JSON output")
    return cmd
}
