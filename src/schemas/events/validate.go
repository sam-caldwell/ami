package events

import "errors"

// Validate performs basic checks on an Event.
// This is a placeholder and will be expanded alongside the spec.
func Validate(e Event) error {
    if e.ID == "" {
        return errors.New("id required")
    }
    return nil
}

