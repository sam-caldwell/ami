package main

import "github.com/spf13/cobra"

// newTestCmd returns the `ami test` subcommand.
func newTestCmd() *cobra.Command {
    var jsonOut bool
    var verbose bool
    var pkgs int
    var checkEvents bool
    var timeoutMs int
    var parallel int
    var failfast bool
    var runPattern string
    var kvMetrics bool
    var kvDump bool
    var kvEvents bool
    cmd := &cobra.Command{
        Use:   "test [path]",
        Short: "Run project tests and write logs/manifests",
        Example: "\n  # Run tests in current directory\n  ami test\n\n  # Run tests in a specific path\n  ami test ./subdir\n\n  # Stream JSON events and summary\n  ami test --json\n\n  # Write test.log and test.manifest under build/test/\n  ami test --verbose\n",
        RunE: func(cmd *cobra.Command, args []string) error {
            dir := "."
            if len(args) > 0 { dir = args[0] }
            if checkEvents { return runCheckEvents(cmd.OutOrStdout()) }
            setTestOptions(TestOptions{TimeoutMs: timeoutMs, Parallel: parallel, Failfast: failfast, RunPattern: runPattern, KvMetrics: kvMetrics, KvDump: kvDump, KvEvents: kvEvents})
            return runTest(cmd.OutOrStdout(), dir, jsonOut, verbose, pkgs)
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON summary (reserved)")
    cmd.Flags().BoolVar(&verbose, "verbose", false, "write build/test artifacts (log and manifest)")
    cmd.Flags().IntVar(&pkgs, "packages", 0, "package-level concurrency for 'go test' (-p); 0 uses default")
    cmd.Flags().BoolVar(&checkEvents, "check-events", false, "internal: exercise events validation")
    _ = cmd.Flags().MarkHidden("check-events")
    cmd.Flags().IntVar(&timeoutMs, "timeout", 0, "default runtime test timeout in milliseconds (0 = no timeout)")
    cmd.Flags().IntVar(&parallel, "parallel", 0, "runtime test concurrency (0 = serial)")
    cmd.Flags().BoolVar(&failfast, "failfast", false, "stop after first failing runtime test")
    cmd.Flags().StringVar(&runPattern, "run", "", "regexp to select runtime tests by name")
    cmd.Flags().BoolVar(&kvMetrics, "kv-metrics", false, "emit kvstore metrics under build/test/kv/")
    cmd.Flags().BoolVar(&kvDump, "kv-dump", false, "emit kvstore dump (keys only) under build/test/kv/")
    cmd.Flags().BoolVar(&kvEvents, "kv-events", false, "stream kv metrics/dump as diag.v1 JSON events in --json mode")
    // alias: pkg-parallel maps to go test -p (same as --packages)
    cmd.Flags().IntVar(&pkgs, "pkg-parallel", 0, "alias for --packages: go test package concurrency (-p)")
    _ = cmd.Flags().MarkHidden("pkg-parallel")
    return cmd
}
