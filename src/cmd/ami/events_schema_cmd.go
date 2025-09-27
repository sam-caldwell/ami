package main

import (
    "fmt"
    "github.com/spf13/cobra"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func newEventsSchemaCmd() *cobra.Command {
    var print bool
    cmd := &cobra.Command{
        Use:    "schema",
        Short:  "Print events.v1 JSON schema (internal)",
        Hidden: true,
        RunE: func(cmd *cobra.Command, args []string) error {
            if print {
                // Write schema to stdout exactly as embedded.
                fmt.Fprint(cmd.OutOrStdout(), ev.SchemaJSON)
                if len(ev.SchemaJSON) == 0 || ev.SchemaJSON[0] != '{' {
                    return fmt.Errorf("events schema not embedded")
                }
                return nil
            }
            // No action without --print to avoid accidental output.
            return nil
        },
    }
    cmd.Flags().BoolVar(&print, "print", false, "print the events.v1 JSON schema")
    return cmd
}

