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

    // Global flags
    cmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "print output as JSON")
    cmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "verbose output")
    cmd.PersistentFlags().BoolVar(&flagColor, "color", false, "colored output (human only)")

    // Wire subcommands
    cmd.AddCommand(cmdInit)
    cmd.AddCommand(cmdClean)
    cmd.AddCommand(cmdLint)
    cmd.AddCommand(cmdTest)
    cmd.AddCommand(cmdBuild)
    cmd.AddCommand(cmdMod)
    cmd.AddCommand(cmdVersion)
    return cmd
}

// Execute runs the root command and returns an exit code.
func Execute() int {
    // reset command-scoped flags (bound to package vars)
    initForce = false
    cmd := newRootCmd()
    if err := cmd.Execute(); err != nil {
        logger.Error(err.Error(), nil)
        return ex.UserError
    }
    return ex.Success
}
