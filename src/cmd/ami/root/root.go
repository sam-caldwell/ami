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

var rootCmd = &cobra.Command{
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

// Execute runs the root command and returns an exit code.
func Execute() int {
    // Global flags
    rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "print output as JSON")
    rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "verbose output")
    rootCmd.PersistentFlags().BoolVar(&flagColor, "color", false, "colored output (human only)")

    // Wire subcommands
    rootCmd.AddCommand(cmdInit)
    rootCmd.AddCommand(cmdClean)
    rootCmd.AddCommand(cmdLint)
    rootCmd.AddCommand(cmdTest)
    rootCmd.AddCommand(cmdBuild)
    rootCmd.AddCommand(cmdMod)
    rootCmd.AddCommand(cmdVersion)

    if err := rootCmd.Execute(); err != nil {
        logger.Error(err.Error(), nil)
        return ex.UserError
    }
    return ex.Success
}

