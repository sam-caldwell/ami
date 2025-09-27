package main

import "github.com/spf13/cobra"

// newBuildCmd returns the `ami build` subcommand (stub).
func newBuildCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "build",
        Short: "Build the workspace (stub)",
        RunE: func(cmd *cobra.Command, args []string) error {
            if lg := getRootLogger(); lg != nil {
                lg.Info("build.start", map[string]any{"args": len(args)})
            }
            // Stub: show help for now to avoid side effects.
            return cmd.Help()
        },
    }
}
