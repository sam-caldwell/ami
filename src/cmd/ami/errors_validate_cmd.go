package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/spf13/cobra"
    es "github.com/sam-caldwell/ami/src/schemas/errors"
)

func newErrorsValidateCmd() *cobra.Command {
    var file string
    cmd := &cobra.Command{
        Use:    "validate",
        Short:  "Validate an errors.v1 JSON file (internal)",
        Hidden: true,
        RunE: func(cmd *cobra.Command, args []string) error {
            if file == "" { return fmt.Errorf("--file is required") }
            b, err := os.ReadFile(file)
            if err != nil { return err }
            var rec es.Error
            if err := json.Unmarshal(b, &rec); err != nil { return err }
            if err := es.Validate(rec); err != nil { return err }
            return nil
        },
    }
    cmd.Flags().StringVar(&file, "file", "", "path to errors.v1 JSON file to validate")
    return cmd
}

