package root

import (
    "os"
    "github.com/spf13/cobra"
)

func newTestCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "test [packages...]",
        Short: "Run Go tests (JSON stream supported)",
        Example: `  ami test ./...
  ami --json test ./...`,
        Run: func(cmd *cobra.Command, args []string) {
            // Ensure flag-bound variables are populated (defensive)
            if v, err := cmd.Flags().GetString("timeout"); err == nil {
                testTimeout = v
            }
            if v, err := cmd.Flags().GetInt("parallel"); err == nil {
                testParallel = v
            }
            if v, err := cmd.Flags().GetInt("pkg-parallel"); err == nil {
                testPkgParallel = v
            }
            if v, err := cmd.Flags().GetBool("failfast"); err == nil {
                testFailFast = v
            }
            if v, err := cmd.Flags().GetString("run"); err == nil {
                testRunFilter = v
            }
            // default to ./...
            if len(args) == 0 {
                args = []string{"./..."}
            }
            code := runGoTests(args)
            // Ensure process exit reflects result for callers
            os.Exit(code)
        },
    }
    // Flags
    cmd.Flags().StringVar(&testTimeout, "timeout", "", "per-package timeout (e.g., 1s, 2m, 10m)")
    cmd.Flags().IntVar(&testParallel, "parallel", 0, "parallelism within package (default GOMAXPROCS)")
    cmd.Flags().IntVar(&testPkgParallel, "pkg-parallel", 0, "number of packages to test in parallel (maps to 'go test -p')")
    cmd.Flags().BoolVar(&testFailFast, "failfast", false, "stop after first test failure")
    cmd.Flags().StringVar(&testRunFilter, "run", "", "run only tests matching regular expression")
    // Runtime tester KV observability flags
    cmd.Flags().BoolVar(&testKVAutoEmit, "kv-metrics", false, "emit kvstore.metrics at end of runtime tests")
    cmd.Flags().BoolVar(&testKVDump, "kv-dump", false, "dump kvstore snapshots at end of runtime tests (human mode)")
    return cmd
}
