package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/spf13/cobra"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func newEventsValidateCmd() *cobra.Command {
    var file string
    cmd := &cobra.Command{
        Use:    "validate",
        Short:  "Validate an events.v1 JSON file (internal)",
        Hidden: true,
        RunE: func(cmd *cobra.Command, args []string) error {
            if file == "" { return fmt.Errorf("--file is required") }
            b, err := os.ReadFile(file)
            if err != nil { return err }
            var e ev.Event
            if err := json.Unmarshal(b, &e); err != nil { return err }
            if err := ev.Validate(e); err != nil { return err }
            return nil
        },
    }
    cmd.Flags().StringVar(&file, "file", "", "path to events.v1 JSON file to validate")
    return cmd
}

