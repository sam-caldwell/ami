package main

import (
    "bytes"
    "fmt"
    "io"
    "os"
    "github.com/spf13/cobra"

    "github.com/sam-caldwell/ami/src/ami/exit"
    "github.com/sam-caldwell/ami/src/ami/logging"
)

// newRootCmd constructs the Cobra root command and attaches subcommands.
func newRootCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:           "ami",
        Short:         "AMI toolchain CLI",
        Long:          "AMI toolchain CLI\n\nExit Codes:\n  0 OK\n  1 Internal error (debug)\n  2 User error (invalid usage/input)\n  3 System I/O error\n  4 Integrity error",
        Example: "\n  # Initialize a new workspace\n  ami init\n\n  # Build with debug artifacts\n  ami build --verbose\n\n  # Lint strictly and emit JSON\n  ami lint --strict --json\n\n  # Run tests and write logs/manifest\n  ami test --verbose\n\n  # Manage modules\n  ami mod list --json\n  ami mod get ./vendor/alpha\n\n  # Visualize pipelines\n  ami pipeline visualize\n",
        SilenceUsage:  true,
        SilenceErrors: true,
        RunE: func(cmd *cobra.Command, args []string) error {
            var buf bytes.Buffer
            // Capture help into a temporary buffer, then write to the original writer.
            w := cmd.OutOrStdout()
            cmd.SetOut(&buf)
            _ = cmd.Help()
            fmt.Fprint(w, buf.String())
            return nil
        },
    }
    // Persistent flags
    cmd.PersistentFlags().Bool("json", false, "enable JSON output")
    cmd.PersistentFlags().Bool("verbose", false, "enable verbose output")
    cmd.PersistentFlags().Bool("color", false, "enable ANSI colors (human output only)")
    cmd.PersistentFlags().StringSlice("redact", nil, "redact these field keys in debug logs (comma-separated)")
    cmd.PersistentFlags().StringSlice("redact-prefix", nil, "redact field keys starting with these prefixes in debug logs")
    cmd.PersistentFlags().StringSlice("allow-field", nil, "allowlist of field keys to include in debug logs")
    cmd.PersistentFlags().StringSlice("deny-field", nil, "denylist of field keys to exclude from debug logs")

    // Validate flag combinations before running any subcommand.
    cmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
        json, _ := cmd.Flags().GetBool("json")
        color, _ := cmd.Flags().GetBool("color")
        verbose, _ := cmd.Flags().GetBool("verbose")
        redact, _ := cmd.Flags().GetStringSlice("redact")
        redactPrefixes, _ := cmd.Flags().GetStringSlice("redact-prefix")
        allowFields, _ := cmd.Flags().GetStringSlice("allow-field")
        denyFields, _ := cmd.Flags().GetStringSlice("deny-field")
        if json && color {
            return exit.New(exit.User, "--json and --color cannot be used together")
        }
        // Construct a root-scoped logger guarded to avoid polluting command outputs.
        // In all modes, we discard primary output; when verbose, also write to build/debug/activity.log.
        opts := logging.Options{
            JSON:     json,
            Verbose:  verbose,
            Color:    color && !json, // colors never in JSON mode
            Package:  "cmd/ami",
            Out:      io.Discard,
            DebugDir: "build/debug",
            RedactKeys:      redact,
            RedactPrefixes:  redactPrefixes,
            FilterAllowKeys: allowFields,
            FilterDenyKeys:  denyFields,
        }
        lg, err := logging.New(opts)
        if err == nil {
            setRootLogger(lg)
        }
        // Ensure package cache directory exists as early as possible per SPEC 1.1.1.1
        if err := ensurePackageCache(); err != nil {
            return err
        }
        return nil
    }

    // Override default help with embedded docs content.
    cmd.SetHelpCommand(newHelpCmd())

    // Register subcommands (keep root stable; each subcommand in its own file).
    cmd.AddCommand(newInitCmd())
    cmd.AddCommand(newCleanCmd())
    cmd.AddCommand(newBuildCmd())
    cmd.AddCommand(newTestCmd())
    cmd.AddCommand(newLintCmd())
    cmd.AddCommand(newVersionCmd())
    // mod commands
    mod := newModCmd()
    mod.AddCommand(newModCleanCmd())
    mod.AddCommand(newModListCmd())
    mod.AddCommand(newModUpdateCmd())
    cmd.AddCommand(mod)
    // pipeline commands
    cmd.AddCommand(newPipelineCmd())
    // events commands (hidden)
    events := newEventsCmd()
    events.AddCommand(newEventsSchemaCmd())
    cmd.AddCommand(events)
    return cmd
}

// execute runs the root command and returns an exit code.
func execute() int {
    root := newRootCmd()
    if err := root.Execute(); err != nil {
        // Ensure error text goes to stderr.
        fmt.Fprintln(os.Stderr, err.Error())
        // Map to exit codes; default to User for unknown errors.
        code := exit.UnwrapCode(err)
        if code == exit.Internal {
            code = exit.User
        }
        return code.Int()
    }
    return 0
}
