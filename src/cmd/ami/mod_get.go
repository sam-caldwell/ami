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
        Example: "\n  # Fetch a local package declared in ami.workspace\n  ami mod get ./vendor/alpha\n\n  # Fetch from git (non-interactive) at tag v1.2.3\n  ami mod get git+ssh://git@github.com/org/repo.git#v1.2.3\n\n  # Emit JSON result\n  ami mod get ./vendor/alpha --json\n",
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
