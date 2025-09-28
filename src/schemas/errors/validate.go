package errors

import "errors"

// Validate performs basic checks on an Error value.
// This is a placeholder and will evolve with the schema and runtime usage.
func Validate(e Error) error {
    if e.Level == "" { return errors.New("level required") }
    if e.Code == "" { return errors.New("code required") }
    if e.Message == "" { return errors.New("message required") }
    return nil
}

