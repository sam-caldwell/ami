package main

import (
    "fmt"
    "github.com/spf13/cobra"
)

// newHelpCmd returns a Cobra help command that prints embedded documentation.
func newHelpCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "help",
        Short: "Show help for ami and subcommands",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Print embedded help documentation.
            _, _ = fmt.Fprintln(cmd.OutOrStdout(), getHelpDoc())
            return nil
        },
    }
}

