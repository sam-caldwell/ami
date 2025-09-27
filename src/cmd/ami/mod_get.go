package main

import (
    "errors"
    "os"
    "github.com/spf13/cobra"
)

// newModGetCmd returns `ami mod get` subcommand.
func newModGetCmd() *cobra.Command {
    var jsonOut bool
    cmd := &cobra.Command{
        Use:   "get <source>",
        Short: "Fetch a module into the local cache",
        Args: func(cmd *cobra.Command, args []string) error {
            if len(args) != 1 {
                return errors.New("expected exactly one source argument")
            }
            return nil
        },
        RunE: func(cmd *cobra.Command, args []string) error {
            wd, err := os.Getwd()
            if err != nil { return err }
            return runModGet(cmd.OutOrStdout(), wd, args[0], jsonOut)
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-parsable JSON output")
    return cmd
}
