package main

import (
    "os"
    "github.com/spf13/cobra"
)

// package-level flags for build options
var buildEmitLLVMOnly bool
var buildNoLink bool
var buildNoLinkEnvs []string
var buildDebugStrict bool
var buildBackend string

// newBuildCmd returns the `ami build` subcommand.
func newBuildCmd() *cobra.Command {
    var jsonOut bool
    cmd := &cobra.Command{
        Use:   "build",
        Short: "Validate workspace and build (phase: validation)",
        Example: "\n  # Human output\n  ami build\n\n  # JSON diagnostics (machine-parsable)\n  ami build --json\n\n  # Write debug artifacts under build/debug/\n  ami build --verbose\n",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Resolve absolute working directory for deterministic behavior.
            wd, err := os.Getwd()
            if err != nil { return err }
            verbose, _ := cmd.Flags().GetBool("verbose")
            return runBuild(cmd.OutOrStdout(), wd, jsonOut, verbose)
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON diagnostics and results")
    cmd.Flags().BoolVar(&buildEmitLLVMOnly, "emit-llvm-only", false, "emit .ll only; skip object compilation")
    cmd.Flags().BoolVar(&buildNoLink, "no-link", false, "skip linking stage (future; reserved)")
    cmd.Flags().StringSliceVar(&buildNoLinkEnvs, "no-link-env", nil, "skip linking for specific target envs (os/arch, comma-separated)")
    cmd.Flags().BoolVar(&buildDebugStrict, "debug-strict", false, "in verbose mode, surface object compile failures as diagnostics")
    cmd.Flags().StringVar(&buildBackend, "backend", "", "codegen backend to use (overrides workspace); e.g., 'llvm'")
    return cmd
}
