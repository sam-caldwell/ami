package main

import "github.com/spf13/cobra"

// newTestCmd returns the `ami test` subcommand (stub).
func newTestCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "test",
        Short: "Run tests (stub)",
        RunE: func(cmd *cobra.Command, args []string) error {
            if lg := getRootLogger(); lg != nil {
                lg.Info("test.start", map[string]any{"args": len(args)})
            }
            return cmd.Help()
        },
    }
}
