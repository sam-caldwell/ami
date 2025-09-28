package main

import (
    "encoding/json"
    "fmt"
    "time"

    "github.com/spf13/cobra"
)

// newPipelineStatsCmd returns `ami pipeline stats` subcommand to expose debug logging pipeline counters.
func newPipelineStatsCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "stats",
        Short: "Show debug logging pipeline counters",
        Example: "\n  # JSON output\n  ami --json pipeline stats\n\n  # Human summary\n  ami pipeline stats\n",
        RunE: func(cmd *cobra.Command, args []string) error {
            lg := getRootLogger()
            if lg == nil {
                // Should not happen as root sets a logger in PersistentPreRunE
                return fmt.Errorf("logger not initialized")
            }
            st, ok := lg.PipelineStats()
            jsonOut, _ := cmd.Root().Flags().GetBool("json")
            if jsonOut {
                out := map[string]any{
                    "schema":    "pipeline.stats.v1",
                    "timestamp": time.Now().UTC().Format(time.RFC3339Nano),
                    "pipeline":  "activity",
                    "active":    ok,
                    "enqueued":  st.Enqueued,
                    "written":   st.Written,
                    "dropped":   st.Dropped,
                    "batches":   st.Batches,
                    "flushes":   st.Flushes,
                }
                return json.NewEncoder(cmd.OutOrStdout()).Encode(out)
            }
            // Human format
            _, _ = fmt.Fprintf(cmd.OutOrStdout(), "pipeline: activity\n")
            _, _ = fmt.Fprintf(cmd.OutOrStdout(), "active: %v\n", ok)
            _, _ = fmt.Fprintf(cmd.OutOrStdout(), "enqueued: %d\n", st.Enqueued)
            _, _ = fmt.Fprintf(cmd.OutOrStdout(), "written: %d\n", st.Written)
            _, _ = fmt.Fprintf(cmd.OutOrStdout(), "dropped: %d\n", st.Dropped)
            _, _ = fmt.Fprintf(cmd.OutOrStdout(), "batches: %d\n", st.Batches)
            _, _ = fmt.Fprintf(cmd.OutOrStdout(), "flushes: %d\n", st.Flushes)
            return nil
        },
    }
    return cmd
}

