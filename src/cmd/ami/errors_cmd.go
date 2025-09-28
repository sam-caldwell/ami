package main

import "github.com/spf13/cobra"

// newErrorsCmd is a hidden parent for errors utilities.
func newErrorsCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:    "errors",
        Short:  "Errors utilities (internal)",
        Hidden: true,
    }
    return cmd
}

