package main

import "github.com/spf13/cobra"

// newTestCmd returns the `ami test` subcommand.
func newTestCmd() *cobra.Command {
    var jsonOut bool
    var verbose bool
    cmd := &cobra.Command{
        Use:   "test [path]",
        Short: "Run project tests and write logs/manifests",
        RunE: func(cmd *cobra.Command, args []string) error {
            dir := "."
            if len(args) > 0 { dir = args[0] }
            return runTest(cmd.OutOrStdout(), dir, jsonOut, verbose)
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON summary (reserved)")
    cmd.Flags().BoolVar(&verbose, "verbose", false, "write build/test artifacts (log and manifest)")
    return cmd
}
