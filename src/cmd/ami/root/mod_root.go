package root

import "github.com/spf13/cobra"

func newModCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "mod",
        Short: "Module and cache operations",
        Example: `  # Clean the module cache
  ami mod clean

  # Fetch a module via git+ssh
  ami mod get git+ssh://git@github.com/org/repo.git#v1.2.3

  # List cached modules (JSON)
  ami mod list --json

  # Update and verify dependencies
  ami mod update
  ami mod verify`,
    }
    cmd.AddCommand(newModCleanCmd())
    cmd.AddCommand(newModUpdateCmd())
    cmd.AddCommand(newModGetCmd())
    cmd.AddCommand(newModListCmd())
    cmd.AddCommand(newModVerifyCmd())
    return cmd
}

