package main

import (
    "fmt"
    "github.com/spf13/cobra"
    errschema "github.com/sam-caldwell/ami/src/schemas/errors"
)

func newErrorsSchemaCmd() *cobra.Command {
    var print bool
    cmd := &cobra.Command{
        Use:    "schema",
        Short:  "Print errors.v1 JSON schema (internal)",
        Hidden: true,
        RunE: func(cmd *cobra.Command, args []string) error {
            if print {
                fmt.Fprint(cmd.OutOrStdout(), errschema.SchemaJSON)
                if len(errschema.SchemaJSON) == 0 || errschema.SchemaJSON[0] != '{' {
                    return fmt.Errorf("errors schema not embedded")
                }
            }
            return nil
        },
    }
    cmd.Flags().BoolVar(&print, "print", false, "print the errors.v1 JSON schema")
    return cmd
}

