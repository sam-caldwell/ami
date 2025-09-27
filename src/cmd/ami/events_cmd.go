package main

import "github.com/spf13/cobra"

// newEventsCmd is a hidden parent for events utilities.
func newEventsCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:    "events",
        Short:  "Events utilities (internal)",
        Hidden: true,
    }
    return cmd
}

