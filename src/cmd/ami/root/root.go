package root

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"

    ex "github.com/sam-caldwell/ami/src/internal/exit"
    "github.com/sam-caldwell/ami/src/internal/logger"
)

var (
    flagJSON    bool
    flagVerbose bool
    flagColor   bool
)

func newRootCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "ami",
        Short: "AMI CLI for workspace, modules, linting, testing, and build",
        Example: `  # Initialize a new workspace
  ami init

  # Clean build artifacts
  ami clean

  # Manage modules
  ami mod get git+ssh://git@github.com/org/repo.git#v1.2.3
  ami mod list --json

  # Build with debug artifacts
  ami build --verbose

  # Print version in JSON
  ami --json version`,
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            // Validate flag interactions
            if flagJSON && flagColor {
                // Plain text to stderr; exit USER_ERROR
                fmt.Fprintln(os.Stderr, "--json and --color cannot be used together")
                os.Exit(ex.UserError)
            }
            logger.Setup(flagJSON, flagVerbose, flagColor)
            return nil
        },
    }
    // Avoid printing usage/errors automatically; tests expect controlled output
    cmd.SilenceUsage = true
    cmd.SilenceErrors = true

    // Global flags
    cmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "print output as JSON")
    cmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "verbose output")
    cmd.PersistentFlags().BoolVar(&flagColor, "color", false, "colored output (human only)")

    // Wire subcommands (fresh instances where constructors exist)
    cmd.AddCommand(newInitCmd())
    cmd.AddCommand(newCleanCmd())
    cmd.AddCommand(newLintCmd())
    cmd.AddCommand(newTestCmd())
    cmd.AddCommand(newBuildCmd())
    cmd.AddCommand(newModCmd())
    cmd.AddCommand(newVersionCmd())
    return cmd
}

// Execute runs the root command and returns an exit code.
func Execute() int {
    // reset command-scoped flags (bound to package vars)
    initForce = false
    cmd := newRootCmd()
    // Snapshot args to avoid cross-package os.Args races during tests.
    if len(os.Args) > 1 {
        args := make([]string, len(os.Args)-1)
        copy(args, os.Args[1:])
        cmd.SetArgs(args)
    }
    if err := cmd.Execute(); err != nil {
        logger.Error(err.Error(), nil)
        return ex.UserError
    }
    return ex.Success
}

// (no adapters)
